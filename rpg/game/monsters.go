package game

type Monster struct {
	Character
}

func NewRat(p Pos) *Monster {
	monster := &Monster{}
	monster.Pos = p
	monster.Name = "Rat"
	monster.Rune = 'R'
	monster.Hitpoints = 10
	monster.Strength = 1
	monster.Speed = 2.0
	monster.ActionPoints = 0.0
	return monster
}

func NewSpider(p Pos) *Monster {
	monster := &Monster{}
	monster.Pos = p
	monster.Name = "Spider"
	monster.Rune = 'S'
	monster.Hitpoints = 10
	monster.Strength = 1
	monster.Speed = 1.0
	monster.ActionPoints = 0.0
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

	//TODO Check it tile being moved to is valid
	//TODO if player is in the way attack
	if !exists && to != level.Player.Pos {
		delete(level.Monsters, m.Pos)
		level.Monsters[to] = m
		m.Pos = to
		return
	}
	if to == level.Player.Pos {
		level.AddEvent(m.Name + " atacked Player!")
		Attack(&m.Character, &level.Player.Character)
	}
}
