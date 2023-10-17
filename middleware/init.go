package middleware

// Init verifies everything is ready to go before application starts
func Init() {
	a := getAuth()
	if a == nil {
		panic("Error initializing AUTH!")
	}

	e := GetEnforcer()
	if e == nil {
		panic("Error initializing RBAC!")
	}
	initEnforcer(e)
}
