package auth

type WhitelistAuther struct {
	whitelist []int64
}

func NewWhiteListAuther(whitelist ...int64) *WhitelistAuther {
	return &WhitelistAuther{
		whitelist: whitelist,
	}
}

func (w WhitelistAuther) AuthN(userID int64) error {
	for _, v := range w.whitelist {
		if userID == v {
			return nil
		}
	}
	return UnauthorizedErr
}
