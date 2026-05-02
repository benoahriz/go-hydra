package main

var routes = Routes{
	Route{"Index", "GET", "/", Index},
	Route{"UploadGet", "GET", "/upload", upload},
	Route{"UploadPost", "POST", "/upload", upload},
	Route{"ConvertPDF", "GET", "/convert/pdf", convertPdfHandler},
	Route{"ToUpperTxt", "GET", "/toupper/txt", toUpperHandler},
	Route{"InvokeFunction", "POST", "/functions/invoke", InvokeFunctionHandler},
	Route{"TodoIndex", "GET", "/todos", TodoIndex},
	Route{"TodoShow", "GET", "/todos/{todoId}", TodoShow},
	Route{"TodoCreate", "POST", "/todos", TodoCreate},
}
