package room

import (
	"backend/util"
	"math/rand"
	"slices"
	"strconv"
	"sync"
)


type Service struct {
	mu        sync.RWMutex
	intervals []util.Interval
	max       int
}

// NewService creates a new RoomService with a given max code value (exclusive).
func NewService(max int) *Service {
	return &Service{
		max: max,
	}
}

func (s *Service) GenerateCode() string {
	paddedLength := 6

	s.mu.Lock()
	defer s.mu.Unlock()

	var code int
	code = s.generateRoomCodeLocked()
	s.insertRoomCodeLocked(code)
	if code >= 100000 {
		return strconv.Itoa(code)
	} else {
		var temp = strconv.Itoa(code)
		for len(temp) < paddedLength {
			temp = "0" + temp
		}
		return temp
	}
}

func (s *Service) generateRoomCodeLocked() int {
	var code int
	code = rand.Intn(s.max)
	if len(s.intervals) == 0 {
		return code
	}
	for i := 0; i < len(s.intervals); i++ {
		start := s.intervals[i].Start
		end := s.intervals[i].End
		if code >= end {
			code += (end - start + 1)
		} else if code >= start {
			code := end + (code - start + 1)
			return code
		} else {
			return code
		}
	}
	s.max = s.max - 1
	return code
}

// Logic behind generating new room code guarantees validity of code.
// Thus, no checks required before inserting new code into intervals.
func (s *Service) insertRoomCodeLocked(code int) {
	var newIntervalIndex = -1
	for i := 0; i < len(s.intervals); i++ {
		start := s.intervals[i].Start
		end := s.intervals[i].End
		if code >= end {
			if code == end + 1 {
				s.intervals[i].End = code
				return
			}
		} else {
			if code+1 == start {
				s.intervals[i].Start = code
				return
			} else {
				newIntervalIndex = i
				break
			}
		}
	}
	// New code larger than all previous ones
	if newIntervalIndex == -1 {
		s.intervals = append(s.intervals, util.Interval{code, code})
	} else {
		newInterval := util.Interval{code, code}
		s.intervals = slices.Insert(s.intervals, newIntervalIndex, newInterval)
	}
}

func (s *Service) VerifyCode(roomCode string) bool {
	var intCode, err = strconv.Atoi(roomCode)
	if err != nil {
		return false
	}
	for i := 0; i < len(s.intervals); i++ {
		if intCode >= s.intervals[i].Start && intCode <= s.intervals[i].End {
			return true
		}
	}
	return false
}