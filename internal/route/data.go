package route

// Data is a convenient template and JSON payload.
//
//	return c.Render("home", route.Data{
//		"Title": "FluxGo",
//		"Admin": true,
//	})
type Data map[string]any

// Set adds a value and returns the same Data for optional fluent construction.
func (data Data) Set(key string, value any) Data {
	data[key] = value
	return data
}

// Merge copies values into Data and returns it.
func (data Data) Merge(values Data) Data {
	for key, value := range values {
		data[key] = value
	}
	return data
}
