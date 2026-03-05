package main

func checkForPipelines(args []string) (position int, hasPipeline bool) {
	for i, arg := range args {
		if arg == "|" {
			return i, true
		}
	}
	return 0, false
}
