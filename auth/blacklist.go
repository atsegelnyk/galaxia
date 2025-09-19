package auth

type BlacklistAuther struct {
	blacklist []int64
}

func NewBlacklistAuther(blacklist ...int64) *BlacklistAuther {
	return &BlacklistAuther{
		blacklist: blacklist,
	}
}

func (b BlacklistAuther) AuthN(userID int64) error {
	for _, v := range b.blacklist {
		if userID == v {
			return UnauthorizedErr
		}
	}
	return nil
}
