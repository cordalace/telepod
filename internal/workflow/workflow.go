package workflow

func NewWorkflow(podRuntime PodRuntime, versionsDB VersionsDB, notifier Notifier) *Workflow {
	return &Workflow{
		podRuntime: podRuntime,
		versionsDB: versionsDB,
		notifier:   notifier,
	}
}

type Workflow struct {
	podRuntime PodRuntime
	versionsDB VersionsDB
	notifier   Notifier
}
