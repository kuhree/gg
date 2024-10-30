package sorts

type Sorter interface {
	Step(arr []int, g *Game) bool // Returns true when sort is complete
	Name() string
}

type QuickSort struct {
	stack [][2]int // Stack for tracking partitions
}

func NewQuickSort() *QuickSort {
	return &QuickSort{
		stack: make([][2]int, 0),
	}
}

func (s *QuickSort) Name() string {
	return "Quick Sort"
}

func (s *QuickSort) Step(arr []int, g *Game) bool {
	if len(s.stack) == 0 {
		if len(arr) > 1 {
			s.stack = append(s.stack, [2]int{0, len(arr) - 1})
		}
		return len(arr) <= 1
	}

	// Pop partition from stack
	last := len(s.stack) - 1
	left, right := s.stack[last][0], s.stack[last][1]
	s.stack = s.stack[:last]

	// Partition
	pivot := arr[right]
	i := left - 1

	for j := left; j < right; j++ {
		g.ComparisonCount++
		if arr[j] <= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
			g.SwapCount++
		}
	}
	arr[i+1], arr[right] = arr[right], arr[i+1]
	pivotIdx := i + 1

	// Push sub-partitions to stack
	if pivotIdx-1 > left {
		s.stack = append(s.stack, [2]int{left, pivotIdx - 1})
	}
	if pivotIdx+1 < right {
		s.stack = append(s.stack, [2]int{pivotIdx + 1, right})
	}

	return len(s.stack) == 0
}

type BubbleSort struct {
	i, j int
}

func NewBubbleSort() *BubbleSort {
	return &BubbleSort{i: 0, j: 0}
}

func (s *BubbleSort) Name() string {
	return "Bubble Sort"
}

func (s *BubbleSort) Step(arr []int, g *Game) bool {
	if s.i >= len(arr)-1 {
		return true
	}

	if s.j >= len(arr)-s.i-1 {
		s.i++
		s.j = 0
		return s.i >= len(arr)-1
	}

	g.ComparisonCount++
	if arr[s.j] > arr[s.j+1] {
		arr[s.j], arr[s.j+1] = arr[s.j+1], arr[s.j]
		g.SwapCount++
	}
	s.j++
	return false
}

type MergeSort struct {
	size    int
	left    int
	tempArr []int
}

func NewMergeSort() *MergeSort {
	return &MergeSort{
		size: 1,
		left: 0,
	}
}

func (s *MergeSort) Name() string {
	return "Merge Sort"
}

func (s *MergeSort) Step(arr []int, g *Game) bool {
	if s.size >= len(arr) {
		return true
	}

	if s.left >= len(arr) {
		s.size *= 2
		s.left = 0
		return s.size >= len(arr)
	}

	mid := s.left + s.size
	if mid > len(arr) {
		mid = len(arr)
	}
	right := mid + s.size
	if right > len(arr) {
		right = len(arr)
	}

	// Create temp array for merging
	temp := make([]int, right-s.left)
	copy(temp, arr[s.left:right])

	// Merge the two halves
	i := 0                 // Index for left half in temp array
	j := mid - s.left     // Index for right half in temp array
	k := s.left           // Index in original array
	rightEnd := right - s.left

	for i < mid-s.left && j < rightEnd {
		g.ComparisonCount++
		if temp[i] <= temp[j] {
			arr[k] = temp[i]
			i++
		} else {
			arr[k] = temp[j]
			j++
		}
		k++
	}

	// Copy remaining elements
	for i < mid-s.left {
		arr[k] = temp[i]
		i++
		k++
	}
	for j < rightEnd {
		arr[k] = temp[j]
		j++
		k++
	}

	s.left += s.size * 2
	return false
}
