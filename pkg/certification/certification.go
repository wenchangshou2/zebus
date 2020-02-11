package certification

type Certification interface {
	Login(params map[string]interface{}) (bool, error)
}
