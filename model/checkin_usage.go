package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

// 签到使用量查询需要索引，请手动在数据库执行：
// CREATE INDEX idx_logs_user_created_at ON logs(user_id, created_at);

// 中国时区，用于计算"昨天"
var chinaTimezone = time.FixedZone("CST", 8*60*60)

// getCheckinNow 统一获取签到逻辑使用的当前时间（东八区）
func getCheckinNow() time.Time {
	return time.Now().In(chinaTimezone)
}

// GetCheckinTimezone 对外暴露签到所用时区，供控制层计算默认月份
func GetCheckinTimezone() *time.Location {
	return chinaTimezone
}

// GetUserYesterdayUsage 获取用户昨天的使用额度（基于中国时区）
func GetUserYesterdayUsage(userId int) (int64, error) {
	now := getCheckinNow()
	yesterday := now.AddDate(0, 0, -1)
	startOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, chinaTimezone)
	endOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, chinaTimezone)

	var totalQuota int64
	err := LOG_DB.Table("logs").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?",
			userId, startOfYesterday.Unix(), endOfYesterday.Unix()).
		Select("COALESCE(SUM(quota), 0)").
		Scan(&totalQuota).Error

	return totalQuota, err
}

// GetUserTodayUsage 获取用户今天的使用额度（基于中国时区）
func GetUserTodayUsage(userId int) (int64, error) {
	now := getCheckinNow()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, chinaTimezone)

	var totalQuota int64
	err := LOG_DB.Table("logs").
		Where("user_id = ? AND created_at >= ?",
			userId, startOfToday.Unix()).
		Select("COALESCE(SUM(quota), 0)").
		Scan(&totalQuota).Error

	return totalQuota, err
}

// CheckYesterdayUsageForCheckin 检查用户昨日使用量是否满足签到条件
// 返回值: (是否满足条件, 昨日使用量, 错误)
func CheckYesterdayUsageForCheckin(userId int) (bool, int64, error) {
	setting := operation_setting.GetCheckinSetting()

	// 如果没有设置最低使用量限制，直接返回通过
	if setting.MinUsageQuota <= 0 {
		return true, 0, nil
	}

	yesterdayUsage, err := GetUserYesterdayUsage(userId)
	if err != nil {
		return false, 0, errors.New("查询使用记录失败")
	}

	if yesterdayUsage < int64(setting.MinUsageQuota) {
		return false, yesterdayUsage, fmt.Errorf("签到需要昨日使用额度达到 %s，您昨日使用额度为 %s",
			logger.LogQuota(setting.MinUsageQuota),
			logger.LogQuota(int(yesterdayUsage)))
	}

	return true, yesterdayUsage, nil
}
