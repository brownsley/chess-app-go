package enum

type GameEndReason string

const (
	ReasonCheckmate     GameEndReason = "checkmate"
	ReasonStalemate     GameEndReason = "stalemate"
	ReasonResignation   GameEndReason = "resignation"
	ReasonTimeOut       GameEndReason = "time_out"
	ReasonDrawAgreement GameEndReason = "draw_agreement"
	ReasonInsufficient  GameEndReason = "insufficient_material"
)
