package mtls

//import "github.com/gofiber/fiber/v2"

type Config struct {
	CaCertificatePath string
}

var DefaultConfig = Config{
	CaCertificatePath: "/etc/peekl/ssl/ca/ca.pem",
}

func defaultConfiguration(config ...Config) {

}

//func New(config Config) fiber.Handler {
//  config :=
//}
