package utils

type ClosingObject interface {
	Close() error
}

func CloseWithoutError(toClose ClosingObject) {
	_ = toClose.Close()
}
