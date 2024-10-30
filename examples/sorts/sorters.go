package sorts

type Sorter interface {
	Step(arr []int) bool // Returns true when sort is complete
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

func (s *QuickSort) Step(arr []int) bool {
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
		if arr[j] <= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
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

func (s *BubbleSort) Step(arr []int) bool {
	if s.i >= len(arr)-1 {
		return true
	}

	if s.j >= len(arr)-s.i-1 {
		s.i++
		s.j = 0
		return s.i >= len(arr)-1
	}

	if arr[s.j] > arr[s.j+1] {
		arr[s.j], arr[s.j+1] = arr[s.j+1], arr[s.j]
	}
	s.j++
	return false
}

type MergeSort struct {
	size     int
	left     int
	stage    int // 0: splitting, 1: merging
	tempArr  []int
}

func NewMergeSort() *MergeSort {
	return &MergeSort{
		size:  1,
		left:  0,
		stage: 0,
	}
}

func (s *MergeSort) Name() string {
	return "Merge Sort"
}

func (s *MergeSort) Step(arr []int) bool {
	if s.size >= len(arr) {
		return true
	}

	if s.left >= len(arr) {
		s.size *= 2
		s.left = 0
		return s.size >= len(arr)
	}

	mid := s.left + s.size
	right := mid + s.size
	if right > len(arr) {
		right = len(arr)
	}

	// Merge
	if s.tempArr == nil {
		s.tempArr = make([]int, right-s.left)
		copy(s.tempArr, arr[s.left:right])
	}

	i, j := s.left, mid
	k := s.left
	for i < mid && j < right {
		if arr[i] <= arr[j] {
			arr[k] = arr[i]
			i++
		} else {
			arr[k] = arr[j]
			j++
		}
		k++
	}

	for i < mid {
		arr[k] = arr[i]
		i++
		k++
	}

	s.left += s.size * 2
	s.tempArr = nil
	return false
}
