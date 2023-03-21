package mach

type ColumnType int

const (
	Int16ColumnType   ColumnType = iota + 0
	Int32ColumnType              = 1
	Int64ColumnType              = 2
	TimeColumnType               = 3
	Float32ColumnType            = 4
	Float64ColumnType            = 5
	IPv4ColumnType               = 6
	Ipv6ColumnType               = 7
	StringColumnType             = 8
	BinaryColumnType             = 9
)

type ColumnSize int

// * DDL: 1-255
// * ALTER SYSTEM: 256-511
// * SELECT: 512
// * INSERT: 513
// * DELETE: 514-517
// * INSERT_SELECT: 518
// * UPDATE: 519
// * EXEC_ROLLUP: 521-523
type StmtType int

func (typ StmtType) IsSelect() bool {
	return typ == 512
}

func (typ StmtType) IsDDL() bool {
	return typ >= 1 && typ <= 255
}

func (typ StmtType) IsAlterSystem() bool {
	return typ >= 256 && typ <= 511
}

func (typ StmtType) IsInsert() bool {
	return typ == 513 || typ == 516
}

func (typ StmtType) IsDelete() bool {
	return typ >= 514 && typ <= 515
}

func (typ StmtType) IsUpdate() bool {
	return typ == 517
}
