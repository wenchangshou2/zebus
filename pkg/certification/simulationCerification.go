package certification

type SimulattionCertification struct{
	Username string
	Password string
}
func (c *SimulattionCertification)Login(params map[string]interface{})(bool,error){
	return true,nil
}