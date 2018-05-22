package object

// Repository provides access to object database.
type Repository interface {
	Put(o *ObjInfo) error
	Get(name string) (*ObjInfo, error)
}
