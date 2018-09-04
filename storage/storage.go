package storage

// InitEnv provides a package level initialization point for any work that is environment specific
func InitEnv(db string) {
	dbURL = db
}
