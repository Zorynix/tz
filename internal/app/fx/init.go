package fx

func ComponentsList() []interface{} {
	return []interface{}{
		LoggerComponent,
		StorageComponent,
		RepositoryComponent,
		ServiceComponent,
		HandlerComponent,
		HTTPComponent,
	}
}

func init() {
	components := ComponentsList()
	for _, c := range components {
		_ = c
	}
}
