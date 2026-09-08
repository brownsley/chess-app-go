package service

import (
	"context"
	"fmt"
	"game-server/internal/game"
	"game-server/internal/request"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

const (
	boardPrefix     = "m:b:"
	playerPrefix    = "m:p:"
	movesPrefix     = "m:m:"
	invitePrefix    = "m:i:"
	matchTypePrefix = "m:t:"
	matchTTL        = 30 * time.Minute
	inviteTTL       = 60 * time.Second
)

func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{
		client: client,
		ctx:    context.Background(),
	}
}

func (s *RedisService) getMatchKeys(matchId string) (boardKey, playerKey, movesKey, matchTypeKey string) {
	return boardPrefix + matchId, playerPrefix + matchId, movesPrefix + matchId, matchTypePrefix + matchId
}

func makeMatchingPrefix(gameMode game.MatchType) string {
	return game.GetQueueName(gameMode)
}

func (s *RedisService) AddToMatchingQueue(gameMode game.MatchType, req request.JoinQueueRequest) error {
	queueKey := makeMatchingPrefix(gameMode)
	memberValue := fmt.Sprintf("%s:%s:%s:%s",
		req.PlayerId,
		req.PlayerName,
		req.Country,
		req.AvatarURL,
	)
	return s.client.ZAdd(s.ctx, queueKey, redis.Z{
		Score:  float64(req.PlayerElo),
		Member: memberValue,
	}).Err()
}

func (s *RedisService) RemoveFromMatchingQueue(gameMode game.MatchType, rawMember string) (bool, error) {
	queueKey := makeMatchingPrefix(gameMode)
	count, err := s.client.ZRem(s.ctx, queueKey, rawMember).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *RedisService) FindMatchingPlayers(gameMode game.MatchType, playerElo int, gap int) ([]redis.Z, error) {
	queueKey := makeMatchingPrefix(gameMode)
	minElo := float64(playerElo - gap)
	maxElo := float64(playerElo + gap)

	args := redis.ZRangeArgs{
		Key:     queueKey,
		ByScore: true,
		Start:   fmt.Sprintf("%f", minElo),
		Stop:    fmt.Sprintf("%f", maxElo),
	}
	return s.client.ZRangeArgsWithScores(s.ctx, args).Result()
}

func (s *RedisService) GetPlayersFromQueue(gameMode game.MatchType, start, end int64) ([]string, error) {
	queueKey := makeMatchingPrefix(gameMode)
	return s.client.ZRange(s.ctx, queueKey, start, end).Result()
}

func (s *RedisService) SetBoardFen(matchId, fen string) error {
	boardKey, playerKey, movesKey, matchTypeKey := s.getMatchKeys(matchId)

	pipe := s.client.Pipeline()
	pipe.Set(s.ctx, boardKey, fen, matchTTL)
	pipe.Expire(s.ctx, playerKey, matchTTL)
	pipe.Expire(s.ctx, movesKey, matchTTL)
	pipe.Expire(s.ctx, matchTypeKey, matchTTL)

	_, err := pipe.Exec(s.ctx)
	return err
}

func (s *RedisService) GetBoardFen(matchId string) (string, error) {
	key := boardPrefix + matchId
	return s.client.Get(s.ctx, key).Result()
}

func (s *RedisService) SetMatchType(matchId, matchType string) error {
	key := matchTypePrefix + matchId
	return s.client.Set(s.ctx, key, matchType, matchTTL).Err()
}

func (s *RedisService) GetMatchType(matchId string) (string, error) {
	key := matchTypePrefix + matchId
	return s.client.Get(s.ctx, key).Result()
}

func (s *RedisService) SetPlayers(matchId string, playersMap map[string]string) error {
	key := playerPrefix + matchId

	values := make(map[string]any, len(playersMap))
	for k, v := range playersMap {
		values[k] = v
	}

	err := s.client.HSet(s.ctx, key, values).Err()
	if err != nil {
		return err
	}
	return s.client.Expire(s.ctx, key, matchTTL).Err()
}

func (s *RedisService) GetPlayers(matchId string) (map[string]string, error) {
	key := playerPrefix + matchId
	return s.client.HGetAll(s.ctx, key).Result()
}

func (s *RedisService) PushMove(matchId, moveNotation string) error {
	boardKey, playerKey, movesKey, matchTypeKey := s.getMatchKeys(matchId)

	err := s.client.RPush(s.ctx, movesKey, moveNotation).Err()
	if err != nil {
		return err
	}

	pipe := s.client.Pipeline()
	pipe.Expire(s.ctx, movesKey, matchTTL)
	pipe.Expire(s.ctx, boardKey, matchTTL)
	pipe.Expire(s.ctx, playerKey, matchTTL)
	pipe.Expire(s.ctx, matchTypeKey, matchTTL)

	_, err = pipe.Exec(s.ctx)
	return err
}

func (s *RedisService) GetMoves(matchId string) ([]string, error) {
	key := movesPrefix + matchId
	return s.client.LRange(s.ctx, key, 0, -1).Result()
}

func (s *RedisService) CleanData(matchId string) error {
	boardKey, playerKey, movesKey, matchTypeKey := s.getMatchKeys(matchId)
	_, err := s.client.Del(s.ctx, boardKey, playerKey, movesKey, matchTypeKey).Result()
	return err
}
