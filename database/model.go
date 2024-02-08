package database

type UsersResponse struct {
	IDUser    uint64 `gorm:"coloumn:id_user"`
	NamaUser  string `gorm:"coloumn:nama_user"`
	IsMutawif int64  `gorm:"coloumn:is_mutawif"`
	IsTl      int64  `gorm:"coloumn:is_tl"`
}

func (UsersResponse) TableName() string {
	return "list_participant_room"
}
