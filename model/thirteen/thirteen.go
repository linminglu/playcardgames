package thirteen

import (
	"encoding/json"
	"fmt"
	dbbill "playcards/model/bill/db"
	enumbill "playcards/model/bill/enum"
	mdbill "playcards/model/bill/mod"
	cacher "playcards/model/room/cache"
	dbr "playcards/model/room/db"
	enumr "playcards/model/room/enum"
	errors "playcards/model/room/errors"
	mdr "playcards/model/room/mod"
	cachet "playcards/model/thirteen/cache"
	dbt "playcards/model/thirteen/db"
	enumt "playcards/model/thirteen/enum"
	errorst "playcards/model/thirteen/errors"
	mdt "playcards/model/thirteen/mod"
	pbt "playcards/proto/thirteen"
	"playcards/utils/db"
	"playcards/utils/log"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/yuin/gopher-lua"
)

//游戏结算逻辑
func CleanGame() []*mdt.GameResultList {
	//从数据库获取未结算的游戏信息
	thirteens, err := GetThirteenByStatusAndGameType()
	if err != nil {
		log.Err("get thirteen by status and game type err :%v", err)
		return nil
	}
	if len(thirteens) == 0 {
		return nil
	}
	//游戏结算结果集合
	var resultListArray []*mdt.GameResultList

	for _, thirteen := range thirteens {
		resultList := &mdt.GameResultList{}
		resultList.RoomID = thirteen.RoomID
		thirteen.Status = enumt.GameStatusDone

		//var resultArray []*mdt.GameResult

		//获取游戏所属房间缓存 更新房间信息
		pwd := cachet.GetRoomPaawordRoomID(thirteen.RoomID)
		room, err := cacher.GetRoom(pwd)
		if err != nil {
			log.Err("room get session failed, %v", err)
			return nil
		}

		var results []*mdt.ThirteenResult
		// var (
		// 	rlist   mdt.GameResultList

		// )

		for _, cards := range thirteen.SubmitCards {
			var shoot = []int32{100003, 100004}
			//var result *mdt.ThirteenResult
			result := &mdt.ThirteenResult{
				UserID: cards.UserID,
				Settle: mdt.ThirteenSettle{
					Head:       "2",
					Middle:     "3",
					Tail:       "4",
					AddScore:   "9",
					TotalScore: "9",
				},
				Result: mdt.ThirteenGroupResult{
					Head: mdt.ResGroup{
						GroupType: "Single",
						CardList:  cards.Head,
					},
					Middle: mdt.ResGroup{
						GroupType: "Couple",
						CardList:  cards.Middle,
					},
					Tail: mdt.ResGroup{
						GroupType: "Couple",
						CardList:  cards.Tail,
					},
				},
				Shoot: shoot,
			}
			results = append(results, result)
		}
		//rlist.RoomID = thirteen.RoomID
		resultList.Result = results

		var resultArray []*mdr.GameUserResult
		for _, result := range resultList.Result {
			m := InitThirteenGameTypeMap()
			for _, userResult := range room.UserResults {
				if userResult.UserID == result.UserID {
					ts, _ := strconv.ParseInt(result.Settle.TotalScore, 10, 32)
					userResult.Score = int32(ts)
					if ts > 0 {
						userResult.Win += 1
					} else if ts == 0 {
						userResult.Tie += 1
					} else if ts == 0 {
						userResult.Lost += 1
					}
					if _, ok := m[result.Result.Head.GroupType]; ok {
						m[result.Result.Head.GroupType]++
					}
					if _, ok := m[result.Result.Middle.GroupType]; ok {
						m[result.Result.Head.GroupType]++
					}
					if _, ok := m[result.Result.Tail.GroupType]; ok {
						m[result.Result.Head.GroupType]++
					}
					m["Shoot"] += int32(len(result.Shoot))
					if len(resultList.Result) > 2 && len(result.Shoot) >= (len(resultList.Result)-1) {
						m["AllShoot"]++
					}
					r, _ := json.Marshal(m)
					userResult.GameCardCount = string(r)
					resultArray = append(resultArray, userResult)
					fmt.Printf("Clean Game userResult Array:%+v", userResult)
				}

			}
		}
		//根据玩家牌组计算结果
		// for _, cards := range thirteen.SubmitCards {
		// 	result := &mdt.GameResult{}
		// 	result.UserID = cards.UserID
		// 	result.CardsList = cards
		// 	var settleArray []*mdt.Settle
		// 	for _, orderCards := range thirteen.SubmitCards {
		// 		var totalScore int32
		// 		if cards.UserID != orderCards.UserID {
		// 			settle := &mdt.Settle{}
		// 			settle.UserID = orderCards.UserID
		// 			settle.ScoreHead = 100
		// 			settle.ScoreMiddle = 150
		// 			settle.ScoreTail = -50
		// 			settle.Score = 200
		// 			settleArray = append(settleArray, settle)
		// 			totalScore += settle.Score
		// 		}

		// 		//更新房间玩家输赢记录
		// 		for _, result := range room.UserResults {
		// 			if result.UserID == cards.UserID {
		// 				result.Score = totalScore
		// 				if totalScore > 0 {
		// 					result.Win += 1
		// 				} else if totalScore == 0 {
		// 					result.Tie += 1
		// 				} else if totalScore == 0 {
		// 					result.Lost += 1
		// 				}
		// 			}
		// 		}

		// 	}

		// 	result.SettleList = settleArray
		// 	result.CardType = "test"
		// 	resultArray = append(resultArray, result)
		// }
		// resultList.Results = resultArray
		resultListArray = append(resultListArray, resultList)
		thirteen.Result = resultList
		//房间局数
		//若到最大局数 则房间流程结束 若没到则重置房间状态和玩家准备状态
		// //room.RoundNow += 1
		// if room.RoundNow >= room.RoundNumber {
		// 	room.Status = enumr.RoomStatusDone
		// } else {
		// 	room.Status = enumr.RoomStatusInit
		// 	for _, user := range room.Users {
		// 		user.Ready = enumr.UserUnready
		// 	}
		// }
		room.Status = enumr.RoomStatusReInit

		//十三张一局结束后庄家顺位移到下一个人
		for i := 0; i < len(room.Users); i++ {
			room.Users[i].Ready = enumr.UserUnready
			if room.Users[i].Role == enumr.UserRoleMaster {
				if i == len(room.Users)-1 {
					room.Users[0].Role = enumr.UserRoleMaster
				} else {
					room.Users[i+1].Role = enumr.UserRoleMaster
				}
			}
		}

		f := func(tx *gorm.DB) error {
			//fmt.Printf("UpdateThirteen:%+v", thirteen)
			thirteen, err = dbt.UpdateThirteen(tx, thirteen)
			if err != nil {
				return err
			}
			r, err := dbr.UpdateRoom(tx, room)
			if err != nil {
				return err
			}
			room = r
			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			log.Err("thirteen update failed, %v", err)
			return nil
		}
		// if room.Status == enumr.RoomStatusDone {
		// 	cacheroom.DeleteRoom(room.Password)
		// } else {
		// 	err = cacheroom.SetRoom(room)
		// }
		err = cacher.SetRoom(room)
		if err != nil {
			log.Err("room create set redis failed,%v | %v", room, err)
			return nil
		}

		err = cachet.DeleteGame(thirteen.RoomID)
		if err != nil {
			log.Err("thirteen set session failed, %v", err)
			return nil
		}
	}

	return resultListArray
}

func InitThirteenGameTypeMap() map[string]int32 {
	m := make(map[string]int32)
	for _, value := range enumt.GroupTypeName {
		m[value] = 0
	}
	return m
}

func CreateThirteen() []*mdt.Thirteen {
	rooms, err := GetRoomsByStatusAndGameType()
	if err != nil {
		log.Err("get rooms by status and game type err :%v", err)
		return nil
	}

	if len(rooms) == 0 {
		return nil
	}
	var newGames []*mdt.Thirteen
	var userResults []*mdr.GameUserResult
	for _, room := range rooms {
		l := lua.NewState()
		defer l.Close()
		if err := l.DoFile("lua/Logic.lua"); err != nil {
			log.Err("thirteen logic do file %v", err)
		}

		if err := l.DoString("return Logic:new()"); err != nil {
			log.Err("thirteen logic do string %v", err)
		}
		//logic := l.Get(1)
		l.Pop(1)
		//fmt.Printf("return value is : %#v\n", ret)
		var groupCards []*mdt.GroupCard
		for _, user := range room.Users {
			if room.RoundNow == 1 {
				userResult := &mdr.GameUserResult{
					UserID: user.UserID,
					Win:    0,
					Lost:   0,
					Tie:    0,
					Score:  0,
				}
				userResults = append(userResults, userResult)
			}
			if err := l.DoString("return Logic:GetCards()"); err != nil {
				log.Err("thirteen logic do string %v", err)
			}
			getCards := l.Get(1)
			l.Pop(1)
			fmt.Printf("return value is : %#v\n", getCards)
			var cardList []string
			if cardsMap, ok := getCards.(*lua.LTable); ok {
				cardsMap.ForEach(func(key lua.LValue, value lua.LValue) {
					if cards, ok := value.(*lua.LTable); ok {
						var cardType string
						var cardValue string
						cards.ForEach(func(k lua.LValue, v lua.LValue) {
							//value, _ := strconv.ParseInt(v.String(), 10, 32)
							if k.String() == "_type" {
								cardType = v.String()
							} else {
								cardValue = v.String()
							}
						})
						// card := mdt.Card{
						// 	Type:  int32(cardType),
						// 	Value: int32(cardValue),
						// }

						cardList = append(cardList, cardType+"_"+cardValue)
					} else {
						log.Err("thirteen cardsMap value err %v", value)
					}
				})
				groupCard := &mdt.GroupCard{
					UserID:   user.UserID,
					CardList: cardList,
				}
				groupCards = append(groupCards, groupCard)
			} else {
				log.Err("thirteen cardsMap err %v", cardsMap)
			}
		}

		thirteen := &mdt.Thirteen{
			RoomID: room.RoomID,
			Status: enumt.GameStatusInit,
			Index:  room.RoundNow,
			//GameLua: l,
			Cards: groupCards,
		}

		f := func(tx *gorm.DB) error {
			err = dbt.CreateThirteen(tx, thirteen)
			if err != nil {
				return err
			}

			room.Status = enumr.RoomStatusStarted
			err = cacher.UpdateRoom(room)
			if err != nil {
				log.Err("room update set session failed, %v", err)
				return err
			}
			_, err = dbr.UpdateRoom(tx, room)

			if room.RoundNow == 1 {
				err := dbbill.GainBalance(tx, room.PayerID,
					&mdbill.Balance{0, 0,
						-int64(room.RoundNumber * enumr.ThirteenGameCost / 10)}, //enumt.GameCost
					enumbill.JournalTypeRoom,
					strconv.Itoa(int(room.GameType))+
						room.Password+
						strconv.Itoa(int(room.RoomID)),
					room.Users[0].UserID, enumbill.DefaultChannel)
				if err != nil {
					return err
				}
				room.UserResults = userResults

				for _, user := range room.Users {
					pr := &mdr.PlayerRoom{
						UserID:    user.UserID,
						RoomID:    room.RoomID,
						GameType:  room.GameType,
						PlayTimes: 0,
					}
					dbr.CreatePlayerRoom(tx, pr)
				}
			}

			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			log.Err("thirteen create failed,%v | %v", thirteen, err)
			continue
		}
		newGames = append(newGames, thirteen)
		cachet.SetGame(thirteen, room.MaxNumber, room.Password)
		if err != nil {
			log.Err("thirteen create set redis failed,%v | %v", room, err)
			continue
		}
		err = cacher.SetRoom(room)
		//fmt.Printf("GameUserResult : %+v\n", room.UserResults)
		if err != nil {
			log.Err("room create set redis failed,%v | %v", room, err)
			continue
		}
		for _, user := range room.Users {
			cachet.SetGameUser(room.RoomID, user.UserID)
		}
	}
	return newGames
}

func SubmitCard(uid int32, submitCard *mdt.SubmitCard) (int32, error) {

	pwd := cacher.GetRoomPasswordByUserID(uid)
	if len(pwd) == 0 {
		return 0, errors.ErrUserNotInRoom
	}
	room, err := cacher.GetRoom(pwd)
	if err != nil {
		return 0, err
	}

	if room.Status > enumr.RoomStatusStarted {
		if room.Status == enumr.RoomStatusWairtGiveUp {
			return 0, errors.ErrInGiveUp
		}
		return 0, errors.ErrGameIsDone
	}

	isReady := cachet.IsGamePlayerReady(room.RoomID, uid)

	if isReady == 0 {
		return 0, errorst.ErrUserNotInGame
	} else if isReady == 2 {
		return 0, errorst.ErrUserAlready
	}

	thirteen, err := cachet.GetGame(room.RoomID)
	if err != nil {
		return 0, err
	}

	submitCard.UserID = uid
	thirteen.SubmitCards = append(thirteen.SubmitCards, submitCard)

	if thirteen.Status > enumt.GameStatusInit {
		return 0, errors.ErrGameIsDone
	}
	playerNow := cachet.GetGamePlayerNowRoomID(room.RoomID)
	playerNow += 1

	if playerNow == room.MaxNumber {
		thirteen.Status = enumt.GameStatusAllReady
	}

	f := func(tx *gorm.DB) error {
		thirteen, err = dbt.UpdateThirteen(tx, thirteen)
		if err != nil {
			return err
		}
		return nil
	}
	err = db.Transaction(f)
	if err != nil {
		return 0, err
	}

	err = cachet.UpdateGameUser(thirteen, uid, playerNow)
	if err != nil {
		log.Err("thirteen set session failed, %v", err)
		return 0, err
	}
	//fmt.Printf("SubmitCardCCCCCCCCC:%+v /n", thirteen)
	return thirteen.RoomID, nil //

}

func Surrender() {}

func GetGameResult() {

}

func GetRoomsByStatusAndGameType() ([]*mdr.Room, error) {
	var (
		rooms []*mdr.Room
	)
	list, err := dbr.GetRoomsByStatusAndGameType(db.DB(),
		enumr.RoomStatusAllReady, enumt.GameID)
	if err != nil {
		return nil, err
	}
	rooms = list
	return rooms, nil
}

func GetThirteenByStatusAndGameType() ([]*mdt.Thirteen, error) {
	var (
		thirteens []*mdt.Thirteen
	)
	list, err := dbt.GetThirteensByStatus(db.DB(),
		enumr.RoomStatusAllReady)
	if err != nil {
		return nil, err
	}
	thirteens = list
	return thirteens, nil
}

func GameResultList(rid int32) (*pbt.GameResultListReply, error) {
	var list []*pbt.GameResultList
	// thirteens, err := dbt.GetThirteenByRoomID(db.DB(), rid)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, thirteen := range thirteens {
	// 	result := thirteen.Result
	// 	list = append(list, result.ToProto())
	// }
	out := &pbt.GameResultListReply{
		List: list,
	}
	return out, nil
}

func CleanGiveUpGame() error {
	var gids []int32
	rids, err := dbr.GetGiveUpRoomIDByGameType(db.DB(), enumt.GameID)
	if err != nil {
		log.Err("get thirteen give up room err:%v", err)
	}

	for _, rid := range rids {
		game, err := cachet.GetGame(rid)
		if err != nil {
			log.Err("get thirteen give up room err:%d|%v", rid, err)
			continue
		}
		if game != nil {
			gids = append(gids, game.GameID)
			err = cachet.DeleteGame(rid)
			if err != nil {
				log.Err(" delete thirteen set session failed, %v", err)
				continue
			}
		}
		fmt.Printf("CleanGiveUpGame:%+v", game)

	}
	if len(gids) > 0 {
		f := func(tx *gorm.DB) error {
			err = dbt.GiveUpGameUpdate(tx, gids)
			if err != nil {
				return err
			}
			return nil
		}
		err = db.Transaction(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetThirteen(rid int32) (*mdt.Thirteen, error) {
	thirteen, err := cachet.GetGame(rid)
	if err != nil {
		return nil, err
	}
	return thirteen, nil
}
