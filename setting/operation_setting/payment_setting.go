package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type PaymentSetting struct {
	AmountOptions  []int           `json:"amount_options"`
	AmountDiscount map[int]float64 `json:"amount_discount"` // 充值金额对应的折扣，例如 100 元 0.9 表示 100 元充值享受 9 折优惠
	AmountBonus    map[int]float64 `json:"amount_bonus"`    // 充值赠送比例，例如 2000 对应 0.2 表示充值 2000 赠送 20%，到账 2400
}

// 默认配置
var paymentSetting = PaymentSetting{
	AmountOptions:  []int{10, 20, 50, 100, 200, 500},
	AmountDiscount: map[int]float64{},
	AmountBonus:    map[int]float64{},
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("payment_setting", &paymentSetting)
}

func GetPaymentSetting() *PaymentSetting {
	return &paymentSetting
}

// GetBonusRate 获取指定充值金额的赠送比例，返回 0 表示无赠送
func GetBonusRate(amount int64) float64 {
	if br, ok := paymentSetting.AmountBonus[int(amount)]; ok && br > 0 {
		return br
	}
	return 0
}
