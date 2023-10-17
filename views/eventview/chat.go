package eventview

import (
	"fmt"
	"os"
	"strings"

	"github.com/tjm/puppet-patching-automation/functions"
	"github.com/tjm/puppet-patching-automation/models"
)

// PrepareMsg prepares a short, markdown-like message which is suitable for sending to chat
// applications like Slack. Formatting like *text* us used to add emphasis. This is supported by
// both Slack and GoogleChat and others. Emoji are also used liberally.
func PrepareMsg(event *models.Event) (msg string) {
	// Message Prefix
	if value, ok := os.LookupEnv("MESSAGE_PREFIX"); ok {
		msg += value
	}

	switch event.Action {
	case models.ActionTest:
		msg += "🧪 Test Message from Patching Automation 🧪"

	// ActionPatchRun*
	case models.ActionPatchRunCreated:
		msg += "❇️ Patch Run Created 🦖\n"
		msg += patchRunDetails(event.PatchRun, event.URL.String())

	case models.ActionPatchRunUpdated:
		msg += "⏫ Patch Run Updated 🦖\n"
		msg += patchRunDetails(event.PatchRun, event.URL.String())

	case models.ActionPatchRunDeleted:
		msg += "❌ " + keyValue("Patch Run Deleted", event.PatchRun.Name)

	// ActionTrelloBoard*
	case models.ActionTrelloBoardCreated:
		board := event.Target.(*models.TrelloBoard)
		msg += "🎯 New Trello Board Created: " + formatLink(board.URL, board.Name)

	case models.ActionTrelloBoardDeleted:
		board := event.Target.(*models.TrelloBoard)
		msg += "❌ " + keyValue("Trello Board Deleted", board.Name)

	// ActionJenkinsBuild*
	case models.ActionJenkinsBuildCreated:
		build := event.Target.(*models.JenkinsBuild)
		msg += "🏗 Jenkins Build Created: " + formatLink(build.URL, build.Name, fmt.Sprintf("#%v", build.APIBuildID))

	// Default just include a message
	default:
		msg += event.Message
	}
	// switch event.GetAction() {
	// case kwrelease.ActionPreInstall:
	// 	msg += fmt.Sprintf("📀 Installing *%s* version *%s* into namespace *%s* via Helm. ⏳\n\nApp version: *%s*\n%s",
	// 		event.GetAppName(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 		event.GetAppVersion(),
	// 		event.GetAppDescription(),
	// 	)

	// case kwrelease.ActionPreUpgrade:
	// 	msg += fmt.Sprintf("⏫ Upgrading *%s* from version %s to version *%s* in namespace *%s* via Helm. ⏳\n%s",
	// 		event.GetAppName(),
	// 		event.GetPreviousChartVersion(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 		getChangeInAppVersion(event),
	// 	)

	// 	if configDiff := getConfigDiff(event); configDiff != "" {
	// 		msg += configDiff
	// 	}

	// case kwrelease.ActionPreRollback:
	// 	msg += fmt.Sprintf("⏬ Rolling back *%s* from version %s to version *%s* in namespace *%s* via Helm. ⏳\n%s",
	// 		event.GetAppName(),
	// 		event.GetPreviousChartVersion(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 		getChangeInAppVersion(event),
	// 	)

	// 	if configDiff := getConfigDiff(event); configDiff != "" {
	// 		msg += configDiff
	// 	}

	// case kwrelease.ActionPreUninstall:
	// 	msg += fmt.Sprintf("🧼 Uninstalling *%s* from namespace *%s* via Helm. ⏳",
	// 		event.GetAppName(),
	// 		event.GetNamespace(),
	// 	)

	// case kwrelease.ActionPostInstall:
	// 	msg += fmt.Sprintf("📀 Installed *%s* version *%s* into namespace *%s* via Helm. ✅\n\n```%s```",
	// 		event.GetAppName(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 		event.GetNotes(),
	// 	)

	// case kwrelease.ActionPostUpgrade:
	// 	msg += fmt.Sprintf("⏫ Upgraded *%s* from version %s to version *%s* in namespace *%s* via Helm. ✅",
	// 		event.GetAppName(),
	// 		event.GetPreviousChartVersion(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 	)

	// case kwrelease.ActionPostRollback:
	// 	msg += fmt.Sprintf("⏬ Rolled back *%s* from version %s to version *%s* in namespace *%s* via Helm. ✅",
	// 		event.GetAppName(),
	// 		event.GetPreviousChartVersion(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 	)

	// case kwrelease.ActionPostReplace:
	// 	msg += fmt.Sprintf("Replaced *%s* version %s with version *%s* in namespace *%s* via Helm. ✅",
	// 		event.GetAppName(),
	// 		event.GetPreviousChartVersion(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 	)

	// case kwrelease.ActionFailedInstall:
	// 	msg += fmt.Sprintf("❌ Installation of *%s* version *%s* in namespace *%s* has FAILED. ❌\n\n```%s```",
	// 		event.GetAppName(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 		// This has the cause of the failure.
	// 		event.GetReleaseDescription(),
	// 	)

	// case kwrelease.ActionFailedReplace:
	// 	msg += fmt.Sprintf("❌ Replacing *%s* version %s with version *%s* in namespace *%s* has FAILED. ❌\n\n```%s```",
	// 		event.GetAppName(),
	// 		event.GetPreviousChartVersion(),
	// 		event.GetChartVersion(),
	// 		event.GetNamespace(),
	// 		event.GetReleaseDescription(),
	// 	)
	// }

	return msg
}

// keyValue will simply return the "key: `value`\n"
func keyValue(key, val string) (msg string) {
	return fmt.Sprintf("**%s:** `%s`\n", key, val)
}

// formatLink will take a url and optionally text and return a string formatted in Google Chat's markdown flavor
// func formatLink(url string, text ...string) string {
// 	if len(text) == 0 {
// 		return "<" + url + ">"
// 	}
// 	return "<" + url + "|" + strings.Join(text, " ") + ">"
// }

// formatLink will take a url and optionally text and return a string formatted in Google Chat's markdown flavor
func formatLink(url string, text ...string) string {
	if len(text) == 0 {
		return "<" + url + ">"
	}
	return "[" + strings.Join(text, " ") + "](" + url + ")"
}

// patchRunDetails will output details of Patch Run
func patchRunDetails(p *models.PatchRun, baseURL string) (msg string) {
	url := fmt.Sprintf("%s/patchRun/%v", baseURL, p.ID)
	msg += "**Name:** " + formatLink(url, p.Name) + "\n"
	if len(p.Description) > 0 {
		msg += keyValue("Description", p.Description)
	}
	msg += keyValue("Start Time", p.StartTime.Format(functions.TimeFormatISO8601))
	msg += keyValue("End Time", p.EndTime.Format(functions.TimeFormatISO8601))
	return
}
