// Package cache @Author  wangjian    2023/6/22 10:27 PM
package cache

type Storage string

const (
	Main  Storage = "main"
	Local Storage = "local"
)

func (s Storage) name() string {
	return string(s)
}
