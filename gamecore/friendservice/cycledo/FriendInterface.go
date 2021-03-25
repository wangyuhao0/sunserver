package cycledo

type FriendInterface struct {
	FS friendService
}

func New(fs friendService) *FriendInterface {
	return &FriendInterface{
		FS: fs,
	}
}

type friendService interface {
}
