package machine

// MachineBuilder builder type for starting a machine
type MachineBuilder struct {
	x *Machine
}

// VertexBuilder builder type for adding a processor to the machine
type VertexBuilder struct {
	m *MachineBuilder
	x *node
}

// RouterBuilder builder type for adding a router to the machine
type RouterBuilder struct {
	m *MachineBuilder
	x *router
}

// CapBuilder builder type for adding capping off a route in the machine
type CapBuilder struct {
	m *MachineBuilder
	x vertex
}

// New func for providing an instance of MachineBuilder
func New(id, name string, fifo bool, i Initium, recorder Recorder) *MachineBuilder {
	return &MachineBuilder{
		x: i.machine(id, name, fifo, recorder),
	}
}

// Build func for providing the underlying machine
func (m *MachineBuilder) Build() *Machine {
	return m.x
}

// Then func for sending the payload to a processor
func (m *MachineBuilder) Then(id, name string, fifo bool, p Processus) *VertexBuilder {
	x := p.convert(id, name, fifo)

	m.x.child = x
	m.x.nodes = map[string]*node{id: x}

	return &VertexBuilder{
		m: m,
		x: x,
	}
}

// Route func for sending the payload to a router
func (m *MachineBuilder) Route(id, name string, fifo bool, r RouteHandler) *RouterBuilder {
	x := r.convert(id, name, fifo)

	m.x.child = x

	return &RouterBuilder{
		m: m,
		x: x,
	}
}

// Cap func for sending the payload to a cap
func (m *MachineBuilder) Cap(id, name string, fifo bool, t Terminus) *CapBuilder {
	x := t.convert(id, name, fifo)

	m.x.child = x

	return &CapBuilder{
		m: m,
		x: x,
	}
}

// To func for sending the payload to a processor
func (m *VertexBuilder) To(v *VertexBuilder) *VertexBuilder {
	m.x.child = v.x
	return v
}

// RouteTo func for sending the payload to a router
func (m *VertexBuilder) RouteTo(r *RouterBuilder) *RouterBuilder {
	m.x.child = r.x
	return r
}

// Then func for sending the payload to a processor
func (m *VertexBuilder) Then(id, name string, fifo bool, p Processus) *VertexBuilder {
	x := p.convert(id, name, fifo)

	m.x.child = x
	m.m.x.nodes[id] = x

	return &VertexBuilder{
		m: m.m,
		x: x,
	}
}

// Route func for sending the payload to a router
func (m *VertexBuilder) Route(id, name string, fifo bool, r RouteHandler) *RouterBuilder {
	x := r.convert(id, name, fifo)

	m.x.child = x

	return &RouterBuilder{
		m: m.m,
		x: x,
	}
}

// Cap func for sending the payload to a cap
func (m *VertexBuilder) Cap(id, name string, fifo bool, t Terminus) *CapBuilder {
	x := t.convert(id, name, fifo)

	m.x.child = x

	return &CapBuilder{
		m: m.m,
		x: x,
	}
}

// ToLeft func for sending the payload to a processor
func (m *RouterBuilder) ToLeft(v *VertexBuilder) *VertexBuilder {
	m.x.left = v.x
	return v
}

// RouteToLeft func for sending the payload to a router
func (m *RouterBuilder) RouteToLeft(r *RouterBuilder) *RouterBuilder {
	m.x.left = r.x
	return r
}

// ThenLeft func for sending the payload to a processor
func (m *RouterBuilder) ThenLeft(id, name string, fifo bool, p Processus) *VertexBuilder {
	x := p.convert(id, name, fifo)

	m.x.left = x
	m.m.x.nodes[id] = x

	return &VertexBuilder{
		m: m.m,
		x: x,
	}
}

// RouteLeft func for sending the payload to a router
func (m *RouterBuilder) RouteLeft(id, name string, fifo bool, r RouteHandler) *RouterBuilder {
	x := r.convert(id, name, fifo)

	m.x.left = x

	return &RouterBuilder{
		m: m.m,
		x: x,
	}
}

// CapLeft func for sending the payload to a cap
func (m *RouterBuilder) CapLeft(id, name string, fifo bool, t Terminus) *CapBuilder {
	x := t.convert(id, name, fifo)

	m.x.left = x

	return &CapBuilder{
		m: m.m,
		x: x,
	}
}

// ToRight func for sending the payload to a processor
func (m *RouterBuilder) ToRight(v *VertexBuilder) *VertexBuilder {
	m.x.left = v.x
	return v
}

// RouteToRight func for sending the payload to a router
func (m *RouterBuilder) RouteToRight(r *RouterBuilder) *RouterBuilder {
	m.x.left = r.x
	return r
}

// ThenRight func for sending the payload to a processor
func (m *RouterBuilder) ThenRight(id, name string, fifo bool, p Processus) *VertexBuilder {
	x := p.convert(id, name, fifo)

	m.x.right = x
	m.m.x.nodes[id] = x

	return &VertexBuilder{
		m: m.m,
		x: x,
	}
}

// RouteRight func for sending the payload to a router
func (m *RouterBuilder) RouteRight(id, name string, fifo bool, r RouteHandler) *RouterBuilder {
	x := r.convert(id, name, fifo)

	m.x.right = x

	return &RouterBuilder{
		m: m.m,
		x: x,
	}
}

// CapRight func for sending the payload to a cap
func (m *RouterBuilder) CapRight(id, name string, fifo bool, t Terminus) *CapBuilder {
	x := t.convert(id, name, fifo)

	m.x.right = x

	return &CapBuilder{
		m: m.m,
		x: x,
	}
}
