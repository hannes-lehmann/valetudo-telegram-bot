package bot

func localizeAttachmentType(attachmentType string) string {
	switch attachmentType {
	case "mop":
		return "Mop"
	case "dustbin":
		return "Dustbin"
	case "watertank":
		return "Watertank"
	default:
		return attachmentType
	}
}

func localizeRobotStatus(robotStatus string) string {
	switch robotStatus {
	case "docked":
		return "Docked"
	case "idle":
		return "Idle"
	case "cleaning":
		return "Cleaning"
	case "paused":
		return "Paused"
	case "returning":
		return "Returning home"
	case "error":
		return "Error"
	case "manual_control":
		return "Manual control"
	}

	return robotStatus
}

func robotStatusEmoji(robotStatus string) string {
	switch robotStatus {
	case "docked":
		return "🏠"
	case "idle":
		return "💤"
	case "cleaning":
		return "🧹"
	case "paused":
		return "⏸"
	case "returning":
		return "🔙"
	case "error":
		return "❗"
	case "manual_control":
		return "🕹"
	}

	return "🤖"
}

func localizeOperationMode(mode string) string {
	switch mode {
	case "vacuum":
		return "Vacuum"
	case "mop":
		return "Mop"
	case "vacuum_and_mop":
		return "Vacuum and mop"
	}

	return mode
}

func operationModeEmoji(mode string) string {
	switch mode {
	case "vacuum":
		return "🧹"
	case "mop":
		return "💧"
	case "vacuum_and_mop":
		return "🧹+💧"
	}

	return mode
}

func localizeFanSpeed(speed string) string {
	switch speed {
	case "low":
		return "Low"
	case "medium":
		return "Medium"
	case "high":
		return "High"
	case "max":
		return "Max"
	}

	return speed
}

func localizeWaterGrade(usage string) string {
	switch usage {
	case "low":
		return "Low"
	case "medium":
		return "Medium"
	case "high":
		return "High"
	}

	return usage
}
