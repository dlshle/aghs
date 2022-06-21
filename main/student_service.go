package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dlshle/aghs/server"
	"github.com/dlshle/aghs/store"
)

type StudentService struct {
	server.Service
	auth         authMiddleware
	studentStore store.AdvancedKVStore
}

const (
	routeStudents     = "/students"
	routeStudentByID  = "/students/:sid"
	routeStudentLogin = "/students/login"
	routeStatus       = "/status"
)

func NewStudentService() StudentService {
	studentService := StudentService{
		nil,
		NewAuthMiddleware(),
		store.NewInMemoryKVStore(200),
	}
	service, _ := server.NewServiceBuilder().
		Id("student").
		WithRouteHandlers(
			server.PathHandlerBuilder(routeStudentByID).
				Get(studentService.handleGetStudent).
				Patch(studentService.handleUpdateStudent).
				Delete(studentService.handleDeleteStudent).
				Build()).
		WithRouteHandlers(
			server.PathHandlerBuilder(routeStudents).
				Get(studentService.handleGetAllStudents).
				Post(studentService.handleAddStudent).
				Build()).
		WithRouteHandlers(
			server.PathHandlerBuilder(routeStudentLogin).
				Post(studentService.handleLogin).
				Build()).
		WithRouteHandlers(server.PathHandlerBuilder(routeStatus).
			Get(server.NewCHandlerBuilder[[]byte]().AddRequiredQueryParam("timestamp").RequireBody().Unmarshaller(func(b []byte) ([]byte, error) {
				return b, nil
			}).OnRequest(func(handle server.CHandle[[]byte]) server.Response {
				return server.NewResponse(http.StatusOK, nil)
			}).MustBuild().HandleRequest).
			Build()).
		Build()
	studentService.Service = service
	return studentService
}

func (s StudentService) handleGetAllStudents(r server.Request) (server.Response, server.ServiceError) {
	students, _ := s.studentStore.Query(func(interface{}) bool {
		return true
	})
	return server.NewResponse(200, students), nil
}

func (s StudentService) handleGetStudent(r server.Request) (server.Response, server.ServiceError) {
	studentId := r.PathParams()["sid"]
	if studentId == "" {
		return nil, server.BadRequestError("invalid student id in path param")
	}
	student, exists := s.getStudentById(studentId)
	if !exists {
		return nil, server.NotFoundError(fmt.Sprintf("student id %s is not found", studentId))
	}
	return server.NewResponse(200, student), nil
}

func (s StudentService) handleAddStudent(r server.Request) (server.Response, server.ServiceError) {
	var toAddStudent Student
	studentData, err := r.Body()
	if err != nil {
		return nil, server.InternalError(err.Error())
	}
	err = json.Unmarshal(studentData, &toAddStudent)
	if err != nil {
		return nil, server.InternalError(err.Error())
	}
	newStudent, err := s.addStudent(toAddStudent)
	if err != nil {
		// only name has already taken error is possible
		return nil, server.BadRequestError(err.Error())
	}
	return server.NewResponse(201, newStudent), nil
}

func (s StudentService) handleUpdateStudent(r server.Request) (server.Response, server.ServiceError) {
	c_sid := r.GetContext(STUDENT_ID_CONTEXT_KEY)
	sid, ok := c_sid.(string)
	if !ok {
		return nil, server.InternalError("unable to get student id from request context")
	}
	toUpdateStudentId := r.PathParams()["sid"]
	if toUpdateStudentId == "" {
		return nil, server.BadRequestError("invalid student id in path param")
	}
	if s.isMyself(sid, toUpdateStudentId) {
		return nil, server.ForbiddenError(fmt.Sprintf("invalid credential to update student %s", toUpdateStudentId))
	}
	updateStudentData, err := r.Body()
	var toUpdateStudent Student
	err = json.Unmarshal(updateStudentData, &toUpdateStudent)
	if err != nil {
		return nil, server.InternalError(err.Error())
	}
	updatedStudent, err := s.updateStudent(toUpdateStudentId, toUpdateStudent)
	if err != nil {
		// 404 is the only case here
		return nil, server.NotFoundError(err.Error())
	}
	return server.NewResponse(200, updatedStudent), nil
}

func (s StudentService) handleDeleteStudent(r server.Request) (server.Response, server.ServiceError) {
	studentId := r.PathParams()["sid"]
	if studentId == "" {
		return nil, server.BadRequestError("invalid student id in path param")
	}
	err := s.deleteStudent(studentId)
	if err != nil {
		return nil, server.BadRequestError(err.Error())
	}
	return server.NewResponse(202, "deleted"), nil
}

type LoginRequest struct {
	Id    string `json:"id"`
	Class string `json:"class"`
}

func (s StudentService) handleLogin(r server.Request) (server.Response, server.ServiceError) {
	loginRequestData, err := r.Body()
	if err != nil {
		return nil, server.InternalError(err.Error())
	}
	var loginRequest LoginRequest
	err = json.Unmarshal(loginRequestData, &loginRequest)
	if err != nil {
		return nil, server.InternalError(err.Error())
	}
	student, exists := s.getStudentById(loginRequest.Id)
	if !exists || student.Class != loginRequest.Class {
		return nil, server.BadRequestError("invalid login credential")
	}
	token := s.auth.grantToken(student)
	return server.NewResponse(200, token), nil
}

func (s StudentService) isMyself(myId, expectedId string) bool {
	return myId == expectedId
}

func (s StudentService) getStudentById(id string) (student Student, exists bool) {
	iStudent, _ := s.studentStore.Get(id)
	if iStudent != nil {
		student = iStudent.(Student)
		exists = true
	}
	return
}

func (s StudentService) addStudent(student Student) (newStudent Student, err error) {
	if exists, _ := s.studentStore.Has(student.Id); exists {
		err = fmt.Errorf("student with name %s already exists", student.Name)
		return
	}
	newStudent = student
	newStudent.Id = newStudent.Name
	s.studentStore.Put(newStudent.Id, newStudent)
	return
}

func (s StudentService) updateStudent(id string, student Student) (updatedStudent Student, err error) {
	stubStudent, exists := s.getStudentById(id)
	if !exists {
		err = fmt.Errorf("student %s does not exist", id)
		return
	}
	if student.Name != "" {
		stubStudent.Name = student.Name
	}
	if student.Class != "" {
		stubStudent.Class = student.Class
	}
	s.studentStore.Put(id, stubStudent)
	updatedStudent = stubStudent
	return
}

func (s StudentService) deleteStudent(id string) error {
	if _, exists := s.getStudentById(id); !exists {
		return fmt.Errorf("student %s does not exit", id)
	}
	s.studentStore.Delete(id)
	return nil
}
