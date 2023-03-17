package hvhstate

type Grade int

const (
	GradeNone Grade = iota - 1
	GradeSub
	GradeMain
)

func (g Grade) String() string {
	switch g {
	case GradeSub:
		return "sub"
	case GradeMain:
		return "main"
	default:
		return ""
	}
}

func (g Grade) IsValid() bool {
	return g >= GradeSub && g <= GradeMain
}

func StringToGrade(name string) Grade {
	switch name {
	case "sub":
		return GradeSub
	case "main":
		return GradeMain
	default:
		return GradeNone
	}
}

type GradeFilter int
const (
	GradeFilterNone GradeFilter = iota - 1
	GradeFilterSub
	GradeFilterMain
	GradeFilterAll
)

func (gf GradeFilter) String() string {
	switch gf {
	case GradeFilterSub:
		return "sub"
	case GradeFilterMain:
		return "main"
	case GradeFilterAll:
		return "all"
	default:
		return ""
	}
}

func StringToGradeFilter(name string) GradeFilter {
	switch name {
	case "sub":
		return GradeFilterSub
	case "main":
		return GradeFilterMain
	case "all":
		return GradeFilterAll
	default:
		return GradeFilterNone
	}
}
