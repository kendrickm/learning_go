package game

type Monster struct {
	Character
}

func NewRat(p Pos) *Monster {
	monster := &Monster{}
	monster.Pos = p
	monster.Name = "Rat"
	monster.Rune = 'R'
	monster.Hitpoints = 5
	monster.Strength = 1
	monster.Speed = 2.0
	monster.ActionPoints = 0.0
	monster.SightRange = 10
	return monster
}

func NewSpider(p Pos) *Monster {
	monster := &Monster{}
	monster.Pos = p
	monster.Name = "Spider"
	monster.Rune = 'S'
	monster.Hitpoints = 20
	monster.Strength = 5
	monster.Speed = 1.0
	monster.ActionPoints = 0.0
	monster.SightRange = 10
	return monster
}

func (m *Monster) Update(level *Level) {
	if m.Hitpoints <= 0 {
		level.AddEvent("You killed the " + m.Name)
		delete(level.Monsters, m.Pos)
		return
	}
	m.ActionPoints += m.Speed
	apInt := int(m.ActionPoints)

	playerPos := level.Player.Pos
	positions := level.astar(m.Pos, playerPos)

	//Do we have a path to the goal?
	if len(positions) == 0 {
		m.Pass()
		return
	}
	moveIndex := 1
	for i := 0; i < apInt; i++ {
		if moveIndex < len(positions) {
			m.Move(positions[moveIndex], level)
			moveIndex++
			m.ActionPoints--
		}

	}
}

func (m *Monster) Pass() {
	m.ActionPoints -= m.Speed
}

func (m *Monster) Move(to Pos, level *Level) {
	_, exists := level.Monsters[to]

	if !exists && to != level.Player.Pos {
		delete(level.Monsters, m.Pos)
		level.Monsters[to] = m
		m.Pos = to
		return
	}
	if to == level.Player.Pos {
		level.Attack(&m.Character, &level.Player.Character)
		if m.Hitpoints <= 0 {
			delete(level.Monsters, m.Pos)
		}
		if level.Player.Hitpoints <= 0 {
			panic("you ded")
		}
	}
}

func (m *Monster) Kill(level *Level) {
	delete(level.Monsters, m.Pos)
	groundItems := level.Items[m.Pos]
	for _, item := range m.Items {
		item.Pos = m.Pos
		groundItems = append(groundItems, item)
	}
	level.Items[m.Pos] = groundItems
}
