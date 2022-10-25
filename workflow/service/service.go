package service

import "tinyflow/workflow/dto"

type Service struct {
	dto dto.DTO
}

func New(dto dto.DTO) Service {
	return Service{dto: dto}
}
