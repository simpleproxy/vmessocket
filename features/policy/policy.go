package policy

type DefaultManager struct{}

func (DefaultManager) Close() error {
	return nil
}

func (DefaultManager) ForSystem() System {
	return System{}
}

func (DefaultManager) Start() error {
	return nil
}

func (DefaultManager) Type() interface{} {
	return ManagerType()
}
