package access

// Grants is the initial role-template policy. Relationships are checked separately.
func Grants(role, capability string) bool {
	switch role {
	case "owner":
		return capability == "organization.manage" || capability == "campus.manage" || capability == "people.manage" || capability == "people.read" || capability == "access.manage" || capability == "audit.read"
	case "coordinator":
		return capability == "campus.manage" || capability == "people.manage" || capability == "people.read"
	default:
		return false
	}
}
