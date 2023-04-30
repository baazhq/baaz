package v1

func (e *Environment) AddCondition(newCon EnvironmentCondition) []EnvironmentCondition {
	for i, c := range e.Status.Conditions {
		if c.Type == newCon.Type {
			if c.Status == newCon.Status {
				return e.Status.Conditions
			}
			e.Status.Conditions[i].Status = newCon.Status
			e.Status.Conditions[i].LastTransitionTime = newCon.LastTransitionTime
			e.Status.Conditions[i].LastUpdateTime = newCon.LastUpdateTime
			e.Status.Conditions[i].Message = newCon.Message
			e.Status.Conditions[i].Reason = newCon.Reason

			return e.Status.Conditions
		}
	}

	e.Status.Conditions = append(e.Status.Conditions, newCon)
	return e.Status.Conditions
}
