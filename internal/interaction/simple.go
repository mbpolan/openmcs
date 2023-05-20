package interaction

// SimpleInterface is an interface with no specific functionality.
type SimpleInterface struct {
	ID int
}

// newSimpleInterface creates a new manager for a no-op interface.
func newSimpleInterface(id int) *SimpleInterface {
	return &SimpleInterface{
		ID: id,
	}
}
