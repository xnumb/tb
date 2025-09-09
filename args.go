package tb

import "github.com/xnumb/tb/to"

type Args map[string]string

func (a Args) Get(k string) string {
	if v, ok := a[k]; ok {
		return v
	}
	return ""
}
func (a Args) GetB(k string) bool {
	if v, ok := a[k]; ok {
		return to.B(v)
	}
	return false
}
func (a Args) GetI(k string) (int, bool) {
	if v, ok := a[k]; ok {
		return to.I(v)
	}
	return 0, false
}
func (a Args) GetI64(k string) (int64, bool) {
	if v, ok := a[k]; ok {
		return to.I64(v)
	}
	return 0, false
}
func (a Args) GetU(k string) (uint, bool) {
	if v, ok := a[k]; ok {
		return to.U(v)
	}
	return 0, false
}
func (a Args) GetF(k string) (float64, bool) {
	if v, ok := a[k]; ok {
		return to.F(v)
	}
	return 0, false
}
