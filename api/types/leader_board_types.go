package types

type LeaderBoardData struct {
	Rank    uint64  `json:"rank" gorm:"column:rank"`
	Address string  `json:"address" gorm:"column:address"`
	Name    string  `json:"name" gorm:"column:name"`
	PnL     float64 `json:"pnl" gorm:"column:pnl"`
	RoI     float64 `json:"roi" gorm:"column:roi"`
}

type GetLeaderBoardResponse []LeaderBoardData
