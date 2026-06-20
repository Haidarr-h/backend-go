package routine

type Service interface {
	CreateRoutine(userID uint, req CreateRoutineRequest) (RoutineResponse, error)
	GetRoutines(userID uint) ([]RoutineResponse, error)
	GetTemplates() ([]RoutineResponse, error)
	GetRoutine(userID uint, routineID uint) (RoutineResponse, error)
	UpdateRoutine(id, userID uint, req UpdateRoutineReq) (RoutineResponse, error)
	DeleteRoutine(routineId uint, userId uint) error
}

type Repo interface {
	Create(routine Routine) (Routine, error)
	FindAll(userID uint) ([]Routine, error)
	FindAllPublic() ([]Routine, error)
	Update(routine Routine, replaceExercises bool) (Routine, error)
	FindByID(id uint) (Routine, error)
	Delete(id uint, userId uint) error
}
