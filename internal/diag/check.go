package diag

type Status int

const (
	StatusUnknown Status = iota
	StatusOK
	StatusNG
	StatusWarn
)

type Result struct {
	Name       string
	NeedsAdmin bool
	Status     Status
	Message    string
	Hint       string
}

type Check interface {
	Name() string
	NeedsAdmin() bool
	Run() Result
}

func RunAll(checks []Check) []Result {
	results := make([]Result, len(checks))
	for i, c := range checks {
		results[i] = runOne(c)
	}
	return results
}

func runOne(c Check) (result Result) {
	defer func() {
		if r := recover(); r != nil {
			result = Result{
				Status:  StatusUnknown,
				Message: "internal error",
				Hint:    "この項目は確認できませんでした。手動で確認してください。",
			}
		}
		result.Name = c.Name()
		result.NeedsAdmin = c.NeedsAdmin()
	}()
	return c.Run()
}
