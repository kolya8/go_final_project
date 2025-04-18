package db

import (
	"database/sql"
	"errors"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TasksResp struct {
	Tasks []*Task `json:"tasks"`
}

func AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?);`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

func Tasks(tasksQuantity int) (*TasksResp, error) {
	tasksResp := &TasksResp{
		Tasks: make([]*Task, 0),
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`
	rows, err := DB.Query(query, tasksQuantity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasksResp.Tasks = append(tasksResp.Tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasksResp, nil
}

func GetTask(id string) (*Task, error) {
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	row := DB.QueryRow(query, id)

	var task Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("task not found")
		}
		return nil, err
	}

	return &task, nil
}

func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("task not found")
	}

	return nil
}

func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE ID = ?`

	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("task not found")
	}

	return nil
}

func UpdateDate(nextDate, id string) error {
	query := `UPDATE scheduler SET Date = ? WHERE ID = ?`
	res, err := DB.Exec(query, nextDate, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("task not found")
	}

	return nil
}

func GetTasksByDate(tasksQuantity int, date string) (*TasksResp, error) {
	tasksResp := &TasksResp{
		Tasks: make([]*Task, 0),
	}
	
	query := "SELECT * FROM scheduler WHERE date = ? LIMIT ?"
	rows, err := DB.Query(query, date, tasksQuantity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasksResp.Tasks = append(tasksResp.Tasks, &task)
	}
	return tasksResp, nil
}

func GetTasksBySearchWord(tasksQuantity int, search string) (*TasksResp, error) {
	tasksResp := &TasksResp{
		Tasks: make([]*Task, 0),
	}

	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ?"
	rows, err := DB.Query(query, "%"+search+"%", "%"+search+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasksResp.Tasks = append(tasksResp.Tasks, &task)
	}
	return tasksResp, nil
}