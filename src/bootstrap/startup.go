package bootstrap

import (
	"zipcode/src/ui"
	"zipcode/src/workspace"
)

func InitialModel(intent StartupIntent) ui.AppModel {
	workspace := workspace.Load(intent.Workspace)

	return ui.AppModel{
		Workspace: workspace,
	}
}
