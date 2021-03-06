package db

import (
	errr "playcards/model/room/errors"
	enum "playcards/model/thirteen/enum"
	mdt "playcards/model/thirteen/mod"
	"playcards/utils/db"
	"playcards/utils/errors"

	"github.com/jinzhu/gorm"
)

func CreateThirteen(tx *gorm.DB, t *mdt.Thirteen) error {
	now := gorm.NowFunc()
	t.UpdatedAt = &now
	// thirteen := &mdt.Thirteen{
	// 	GameResults: t.GameResults,
	// 	Status:      t.Status,
	// 	UpdatedAt:   &now,
	// }
	if err := tx.Create(t).Error; err != nil {
		return errors.Internal("create thirteen failed", err)
	}
	return nil
}

func UpdateThirteen(tx *gorm.DB, t *mdt.Thirteen) (*mdt.Thirteen, error) {
	now := gorm.NowFunc()
	thirteen := &mdt.Thirteen{
		UserSubmitCards: t.UserSubmitCards,
		GameResults:     t.GameResults,
		Status:          t.Status,
		UpdatedAt:       &now,
	}
	if err := tx.Model(t).Updates(thirteen).Error; err != nil {
		return nil, errors.Internal("update thirteen failed", err)
	}
	return t, nil
}

func GetThirteensByStatus(tx *gorm.DB, status int32) ([]*mdt.Thirteen, error) {
	var (
		out []*mdt.Thirteen
	)
	if err := tx.Where("status = ?", status).Order("created_at").
		Find(&out).Error; err != nil {
		return nil, errr.ErrRoomNotExisted
	}
	return out, nil
}

func GetThitteenByID(tx *gorm.DB, gid int32) (*mdt.Thirteen, error) {
	var (
		out mdt.Thirteen
	)
	out.GameID = gid
	found, err := db.FoundRecord(tx.Find(&out).Error)
	if err != nil {
		return nil, errors.Internal("get thirteen failed", err)
	}

	if !found {
		return nil, errr.ErrRoomNotExisted
	}
	return &out, nil
}

func GetThirteenByRoomID(tx *gorm.DB, rid int32) ([]*mdt.Thirteen, error) {
	var out []*mdt.Thirteen
	if err := tx.Where(" room_id = ? ", rid).
		Order("created_at").Find(&out).Error; err != nil {
		return nil, errors.Internal("get thirteen by room_id failed", err)
	}
	return out, nil
}

func GetLastThirteenByRoomID(tx *gorm.DB, rid int32) (*mdt.Thirteen, error) {
	out := &mdt.Thirteen{}
	if err := tx.Where(" room_id = ? ", rid).
		Order("game_id desc").Limit(1).Find(&out).Error; err != nil {
		return nil, errors.Internal("get last thirteen by room_id failed", err)
	}
	//if err := tx.Where(" room_id = ? and index = ?", rid,index).
	//Error; err != nil {
	//	return nil, errors.Internal("get last thirteen by room_id failed", err)
	//}
	return out, nil
}

func GiveUpGameUpdate(tx *gorm.DB, gids []int32) error {
	if err := tx.Table(enum.ThirteenTableName).Where(" game_id IN (?)", gids).
		Updates(map[string]interface{}{"status": enum.GameStatusGiveUp}).
		Error; err != nil {
		return errors.Internal("get thirteen by room_id failed", err)
	}
	return nil
}
