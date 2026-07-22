package config

type Captcha struct {
	KeyLong            int `mapstructure:"key-long" json:"key-long" yaml:"key-long"`                                     // CAPTCHA length
	ImgWidth           int `mapstructure:"img-width" json:"img-width" yaml:"img-width"`                                  // CAPTCHA width
	ImgHeight          int `mapstructure:"img-height" json:"img-height" yaml:"img-height"`                               // CAPTCHA height
	OpenCaptcha        int `mapstructure:"open-captcha" json:"open-captcha" yaml:"open-captcha"`                         // This number enables anti-brute-force CAPTCHA. 0 means a CAPTCHA is required on every login, while other numbers represent incorrect password attempts. For example, 3 means a CAPTCHA will appear after three incorrect attempts.
	OpenCaptchaTimeOut int `mapstructure:"open-captcha-timeout" json:"open-captcha-timeout" yaml:"open-captcha-timeout"` // Anti-brute-force CAPTCHA timeout, unit: seconds (s).
}
