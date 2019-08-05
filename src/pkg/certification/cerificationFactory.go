package certification

import "github.com/wenchangshou2/zebus/src/pkg/setting"

var (
	G_Certification Certification
)

func InitCertification() error {
	model := setting.ServerSetting.AuthModel
	switch model {
	case "local":
		username := setting.ServerSetting.AuthUsername
		password := setting.ServerSetting.AuthPassword
		G_Certification = &LocalCertification{
			Username: username,
			Password: password,
		}
	default:
		G_Certification = &SimulattionCertification{}
	}
	return nil
}
