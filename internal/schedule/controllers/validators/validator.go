package validators

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"studyum/internal/schedule/dto"
)

var ValidationError = errors.New("validation error")

type Validator interface {
	AddGeneralLesson(dto dto.AddGeneralLessonDTO) error

	AddLesson(dto dto.AddLessonDTO) error
	UpdateLesson(lessonDTO dto.UpdateLessonDTO) error
}

type schedule struct {
	validate *validator.Validate
}

func NewSchedule(validate *validator.Validate) Validator {
	return &schedule{validate: validate}
}

func (s *schedule) AddGeneralLesson(dto dto.AddGeneralLessonDTO) error {
	if dto.StartTimeMinutes <= dto.EndTimeMinutes {
		return errors.Wrap(ValidationError, "start time is after end time")
	}

	return nil
}

func (s *schedule) AddLesson(dto dto.AddLessonDTO) error {
	if !dto.StartDate.Before(dto.EndDate) {
		return errors.Wrap(ValidationError, "start date is after end date")
	}

	return nil
}

func (s *schedule) UpdateLesson(dto dto.UpdateLessonDTO) error {
	if err := s.AddLesson(dto.AddLessonDTO); err != nil {
		return err
	}

	if dto.Id.IsZero() {
		return errors.Wrap(ValidationError, "not valid id")
	}

	return nil
}
