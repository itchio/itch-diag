package main

const baseHTML = `
	<!doctype html>
	<html>
		<head>
			<title>itch diagnostics</title>
			<style>
			* {
				box-sizing: border-box;
				font-family: sans-serif;
			}

			body {
				padding: 10px 20px;
				overflow-y: scroll;
			}
			
			i {
				font-variant: italic;
			}
			</style>
		</head>

		<body>
			<div id="app"></div>
		</body>
	</html>
`
