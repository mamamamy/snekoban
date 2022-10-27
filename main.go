package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	CELL_WALL     = "wall"
	CELL_TARGET   = "target"
	CELL_COMPUTER = "computer"
	CELL_PLAYER   = "player"
	CELL_NULL     = ""
)

var DIRECTION_VECTOR = map[string][2]int{
	"up":    {-1, 0},
	"down":  {+1, 0},
	"left":  {0, -1},
	"right": {0, +1},
}

var DIRECTION_INT = map[string]uint8{
	"up":    0,
	"down":  1,
	"left":  2,
	"right": 3,
}

var DIRECTION_LIST = []string{
	"up",
	"down",
	"left",
	"right",
}

type DebugDataStruct []any

func (x *DebugDataStruct) Push(data any) {
	*x = append(*x, fmt.Sprint(data))
}

var DebugData DebugDataStruct

type QueueNode struct {
	next *QueueNode
	prev *QueueNode
	data any
}

type Queue struct {
	head *QueueNode
	tail *QueueNode
}

func (q *Queue) Empty() bool {
	return q.head == nil
}

func (q *Queue) Push(x any) {
	newNode := &QueueNode{
		data: x,
	}
	if q.tail != nil {
		q.tail.next = newNode
	}
	newNode.prev = q.tail
	q.tail = newNode
	if q.head == nil {
		q.head = newNode
	}
}

func (q *Queue) Pop() any {
	if q.Empty() {
		panic(errors.New("q.Empty()"))
	}
	node := q.head
	q.head = node.next
	if q.head == nil {
		q.tail = nil
	}
	return node.data
}

type SolveState struct {
	State    GameState
	StepList []uint8
}

type SolveStateQueue struct {
	Queue
}

func (q *SolveStateQueue) Push(x SolveState) {
	q.Queue.Push(x)
}

func (q *SolveStateQueue) Pop() SolveState {
	return q.Queue.Pop().(SolveState)
}

type Vec2 struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Vec2Set map[Vec2]struct{}

func (v *Vec2Set) UnmarshalJSON(data []byte) error {
	*v = map[Vec2]struct{}{}
	tmp := map[string]struct{}{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	for k := range tmp {
		vec2 := Vec2{}
		json.Unmarshal([]byte(k), &vec2)
		(*v)[vec2] = struct{}{}
	}
	return nil
}

func (v Vec2Set) MarshalJSON() ([]byte, error) {
	tmp := map[string]struct{}{}
	for k := range v {
		s, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		tmp[string(s)] = struct{}{}
	}
	return json.Marshal(tmp)
}

type GameState struct {
	StaticLayer [][]string `json:"staticLayer"`
	ComputerSet Vec2Set
	Player      Vec2 `json:"player"`
}

func NewGameState(description [][][]string) GameState {
	gs := GameState{
		ComputerSet: map[Vec2]struct{}{},
	}
	gs.StaticLayer = make([][]string, len(description))
	for i, v := range description {
		gs.StaticLayer[i] = make([]string, len(v))
		for j, v := range v {
			for _, v := range v {
				switch v {
				case CELL_TARGET:
					fallthrough
				case CELL_WALL:
					gs.StaticLayer[i][j] = v
				case CELL_COMPUTER:
					gs.ComputerSet[Vec2{X: j, Y: i}] = struct{}{}
				case CELL_PLAYER:
					gs.Player = Vec2{X: j, Y: i}
				}
			}
		}
	}
	return gs
}

func (gs *GameState) DumpGame() [][][]string {
	description := make([][][]string, len(gs.StaticLayer))
	for i, v := range gs.StaticLayer {
		description[i] = make([][]string, len(v))
		for j, v := range v {
			description[i][j] = []string{}
			if v != CELL_NULL {
				description[i][j] = append(description[i][j], v)
			}
		}
	}
	for k := range gs.ComputerSet {
		description[k.Y][k.X] = append(description[k.Y][k.X], CELL_COMPUTER)
	}
	description[gs.Player.Y][gs.Player.X] = append(description[gs.Player.Y][gs.Player.X], CELL_PLAYER)
	return description
}

func (gs *GameState) VictoryCheck() bool {
	if len(gs.ComputerSet) <= 0 {
		return false
	}
	for k := range gs.ComputerSet {
		if gs.StaticLayer[k.Y][k.X] != CELL_TARGET {
			return false
		}
	}
	return true
}

func (gs *GameState) StepGame(direction string) GameState {
	directionVector := DIRECTION_VECTOR[direction]
	player := gs.Player
	player.Y += directionVector[0]
	player.X += directionVector[1]
	if gs.StaticLayer[player.Y][player.X] == CELL_WALL {
		return *gs
	}
	computer := Vec2{X: player.X, Y: player.Y}
	_, ok := gs.ComputerSet[computer]
	if ok {
		computerFront := computer
		computerFront.Y += directionVector[0]
		computerFront.X += directionVector[1]
		if gs.StaticLayer[computerFront.Y][computerFront.X] == CELL_WALL {
			return *gs
		}
		_, ok := gs.ComputerSet[computerFront]
		if ok {
			return *gs
		}
		delete(gs.ComputerSet, computer)
		gs.ComputerSet[computerFront] = struct{}{}
	}
	gs.Player = player
	return *gs
}

func (gs *GameState) Copy() GameState {
	newGs := *gs
	newGs.ComputerSet = map[Vec2]struct{}{}
	for k := range gs.ComputerSet {
		newGs.ComputerSet[k] = struct{}{}
	}
	return newGs
}

func (gs *GameState) StateKey() string {
	unDupKey := fmt.Sprint(gs.ComputerSet, gs.Player)
	unDupKey = strings.ReplaceAll(unDupKey, "map[{", "")
	unDupKey = strings.ReplaceAll(unDupKey, "}:{}] {", ",")
	unDupKey = strings.ReplaceAll(unDupKey, "}", "")
	unDupKey = strings.ReplaceAll(unDupKey, " ", ",")
	return unDupKey
}

func (gs *GameState) SolvePuzzle() []string {
	if gs.VictoryCheck() {
		return []string{}
	}
	q := SolveStateQueue{}
	q.Push(SolveState{
		State:    gs.Copy(),
		StepList: []uint8{},
	})
	unDupSet := map[string]struct{}{}
	unDupSet[gs.StateKey()] = struct{}{}
	for !q.Empty() {
		currentState := q.Pop()
		for direction := range DIRECTION_VECTOR {
			newState := currentState.State.Copy()
			newState.StepGame(direction)
			if newState.VictoryCheck() {
				stepList := make([]string, len(currentState.StepList))
				for i, v := range currentState.StepList {
					stepList[i] = DIRECTION_LIST[v]
				}
				stepList = append(stepList, direction)
				DebugData.Push(len(stepList))
				return stepList
			}
			unDupKey := newState.StateKey()
			_, ok := unDupSet[unDupKey]
			if !ok {
				stepList := []uint8{}
				stepList = append(stepList, currentState.StepList...)
				stepList = append(stepList, DIRECTION_INT[direction])
				q.Push(SolveState{
					State:    newState,
					StepList: stepList,
				})
				unDupSet[unDupKey] = struct{}{}
			}
		}
	}
	return nil
}

func doNewGame(data []byte) any {
	var description [][][]string
	err := json.Unmarshal(data, &description)
	if err != nil {
		panic(err)
	}
	return NewGameState(description)
}

func doDumpGame(data []byte) any {
	var gs GameState
	err := json.Unmarshal(data, &gs)
	if err != nil {
		panic(err)
	}
	return gs.DumpGame()
}

func doVictoryCheck(data []byte) any {
	var gs GameState
	err := json.Unmarshal(data, &gs)
	if err != nil {
		panic(err)
	}
	return gs.VictoryCheck()
}

type StepGameData struct {
	Game      GameState `json:"game"`
	Direction string    `json:"direction"`
}

func doStepGame(data []byte) any {
	var sgd StepGameData
	err := json.Unmarshal(data, &sgd)
	if err != nil {
		panic(err)
	}
	return sgd.Game.StepGame(sgd.Direction)
}

func doSolvePuzzle(data []byte) any {
	var gs GameState
	err := json.Unmarshal(data, &gs)
	if err != nil {
		panic(err)
	}
	return gs.SolvePuzzle()
}

func doOutput(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, bytes.NewReader(b))
}

func main() {
	defer func() {
		err := recover().(error)
		if err != nil {
			doOutput(map[string]any{"errCode": 1, "errMsg": err.Error(), "DEBUG_DATA": DebugData})
		}
	}()
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	doMap := map[string]func(data []byte) any{
		"-new_game":      doNewGame,
		"-victory_check": doVictoryCheck,
		"-step_game":     doStepGame,
		"-dump_game":     doDumpGame,
		"-solve_puzzle":  doSolvePuzzle,
	}
	doOutput(map[string]any{"errCode": 0, "data": doMap[os.Args[1]](data), "DEBUG_DATA": DebugData})
}
