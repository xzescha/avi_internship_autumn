package domain

// BulkDeactivateResult — результат массовой деактивации пользователей команды.
type BulkDeactivateResult struct {
	TeamName         string
	DeactivatedUsers int64 // сколько реально стало неактивными в этой операции
	AffectedPRs      int   // сколько открытых PR пришлось трогать (удалять/менять ревьюверов)
}
