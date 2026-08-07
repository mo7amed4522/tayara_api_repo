package payment

import (
	"strconv"

	"github.com/mo7amed4522/tayara_api/constants"
	"github.com/mo7amed4522/tayara_api/global"
	baseModel "github.com/mo7amed4522/tayara_api/model/base"
)

// stripeFeeBPS returns the current Stripe fee in basis points. The runtime System Setting
// (`stripe_fee_percent`) wins over the deploy-time `config.Stripe.FeeBPS` default, so admin can
// change the fee without redeploying. Falls back to config when the setting is absent/invalid.
func stripeFeeBPS() int64 {
	if global.GVA_DB == nil {
		return int64(global.GVA_CONFIG.Stripe.FeeBPS)
	}
	var st baseModel.SystemSetting
	err := global.GVA_DB.Where("key = ?", constants.SettingKeyStripeFeePercent).First(&st).Error
	if err != nil {
		// fall back to the config default
		return int64(global.GVA_CONFIG.Stripe.FeeBPS)
	}
	if v, convErr := strconv.ParseInt(st.Value, 10, 64); convErr == nil && v >= 0 {
		return v
	}
	return int64(global.GVA_CONFIG.Stripe.FeeBPS)
}
