// Package util @Author  wangjian    2023/6/23 12:53 PM
package util

import "time"

func MinDuration(x, y time.Duration) time.Duration {
	if x < y {
		return x
	}
	return y
}

func MaxDuration(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}
	return y
}
