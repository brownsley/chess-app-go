package service

import (
	"context"
	"fmt"
	"game-server/internal/game"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

const (
	boardPrefix     = "match:board:"
	playerPrefix    = "match:players:"
	movesPrefix     = "match:moves:"
	invitePrefix    = "match:invite:"
	matchTypePrefix = "match:type:"
	matchTTL        = 2 * time.Hour
	inviteTTL       = 60 * time.Second
)

func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{
		client: client,
		ctx:    context.Background(),
	}
}

func makeMatchingPrefix(gameMode game.MatchType) string {
	return game.GetQueueName(gameMode)
}

func (s *RedisService) AddToMatchingQueue(gameMode game.MatchType, playerId string, playerElo int) error {
	queueKey := makeMatchingPrefix(gameMode)
	return s.client.ZAdd(s.ctx, queueKey, redis.Z{
		Score:  float64(playerElo),
		Member: playerId,
	}).Err()
}

func (s *RedisService) RemoveFromMatchingQueue(gameMode game.MatchType, playerId string) error {
	queueKey := makeMatchingPrefix(gameMode)
	return s.client.ZRem(s.ctx, queueKey, playerId).Err()
}

func (s *RedisService) FindMatchingPlayers(gameMode game.MatchType, playerElo int, gap int) ([]string, error) {
	queueKey := makeMatchingPrefix(gameMode)
	minElo := float64(playerElo - gap)
	maxElo := float64(playerElo + gap)

	args := redis.ZRangeArgs{
		Key:     queueKey,
		ByScore: true,
		Start:   fmt.Sprintf("%f", minElo),
		Stop:    fmt.Sprintf("%f", maxElo),
	}
	return s.client.ZRangeArgs(s.ctx, args).Result()
}

func (s *RedisService) GetPlayersFromQueue(gameMode game.MatchType, start, end int64) ([]string, error) {
	queueKey := makeMatchingPrefix(gameMode)
	return s.client.ZRange(s.ctx, queueKey, start, end).Result()
}

func (s *RedisService) SetBoardFen(matchId, fen string) error {
	key := boardPrefix + matchId
	return s.client.Set(s.ctx, key, fen, matchTTL).Err()
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
	key := movesPrefix + matchId
	err := s.client.RPush(s.ctx, key, moveNotation).Err()
	if err != nil {
		return err
	}
	return s.client.Expire(s.ctx, key, matchTTL).Err()
}

func (s *RedisService) GetMoves(matchId string) ([]string, error) {
	key := movesPrefix + matchId
	return s.client.LRange(s.ctx, key, 0, -1).Result()
}

func (s *RedisService) CleanData(matchId string) error {
	boardKey := boardPrefix + matchId
	playerKey := playerPrefix + matchId
	movesKey := movesPrefix + matchId
	matchTypeKey := matchTypePrefix + matchId

	_, err := s.client.Del(s.ctx, boardKey, playerKey, movesKey, matchTypeKey).Result()
	return err
}
