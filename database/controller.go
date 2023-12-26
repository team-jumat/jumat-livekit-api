package database

func GetUserById(id int64) (*UsersResponse, error) {
	db, _ := GetInstance()

	var data UsersResponse
	if err := db.First(&data, "id_user=?", id).Error; err != nil {
		return nil, err
	}

	return &data, nil
}
