package models

type NullString string

func (ns *NullString) Scan(value interface{}) error {
	if source, ok := value.(string); ok {
		*ns = NullString(source)
	} else {
		*ns = ""
	}

	return nil
}
