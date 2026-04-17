package main

func (a *App) doCd(args []string) {
	if len(args) == 0 {
		a.cwd = "/" + a.homeChannel
		return
	}

	target := args[0]
	switch target {
	case "..":
		chID, plID := parsePath(a.cwd)
		switch {
		case plID != "":
			a.cwd = "/" + chID
		case chID != "":
			a.cwd = "//"
		}
	case "~":
		a.cwd = "/" + a.homeChannel
	case "/", "//":
		a.cwd = "//"
	default:
		chID, plID := parsePath(a.cwd)
		if chID == "" {
			a.cwd = "/" + target
		} else if plID == "" {
			a.cwd = "/" + chID + "/" + target
		}
	}
}
