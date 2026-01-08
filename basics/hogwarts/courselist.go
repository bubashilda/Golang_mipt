//go:build !solution

package hogwarts

type DFSState int

const (
	WHITE DFSState = iota
	GREY
	BLACK
)

func DFSDependencies(prereq string, prereqs map[string][]string, visited map[string]DFSState) []string {
	var ans []string

	if visited[prereq] == BLACK {
		return ans
	}
	if visited[prereq] == GREY {
		panic("cycle in deps graph")
	}
	visited[prereq] = GREY

	for _, item := range prereqs[prereq] {
		ans = append(ans, DFSDependencies(item, prereqs, visited)...)
	}

	visited[prereq] = BLACK
	ans = append(ans, prereq)
	return ans
}

func GetCourseList(prereqs map[string][]string) []string {
	var ans []string
	visited := make(map[string]DFSState)

	for key := range prereqs {
		if visited[key] == WHITE {
			ans = append(ans, DFSDependencies(key, prereqs, visited)...)
		}
	}

	return ans
}
