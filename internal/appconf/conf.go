package appconf

type Config struct {
	Listen       string `default:":4560"`
	Debug        bool   `default:"false"`
	Base         string `required:"true"`
	CloudreveKey string `required:"true" split_words:"true"`

	EpayPartnerID    string `required:"true" split_words:"true"`
	EpayKey          string `required:"true" split_words:"true"`
	EpayEndpoint     string `required:"true" split_words:"true"`
	EpayPurchaseType string `default:"alipay" split_words:"true"`

	RedisEnabled  bool   `default:"false" split_words:"true"`
	RedisServer   string `default:"localhost:6379" split_words:"true"`
	RedisPassword string `default:"" split_words:"true"`
	RedisDB       int    `default:"0" split_words:"true"`

	CustomName string `default:"" split_words:"true"`
	AutoSubmit bool   `default:"true" split_words:"true"`

	// USDTMore 配置
	USDTMoreEnabled      bool   `default:"false" split_words:"true"`
	USDTMoreAPIEndpoint  string `default:"http://localhost:6080" split_words:"true"`
	USDTMoreAuthToken    string `default:"123456" split_words:"true"`
	USDTMoreDefaultChain string `default:"TRON" split_words:"true"`
}
