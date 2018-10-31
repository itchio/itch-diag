package main

const baseCSS = `
	* {
		box-sizing: border-box;
		font-family: Lato, sans-serif;
		line-height: 1.4;
		font-size: 13px;
	}

	body {
		background: #1d1c1c;
		color: white;
		padding: 5px 10px;
		overflow-y: scroll;
	}

	p {
		margin: 0.1em 0;
		padding: 0;
	}

	pre {
		margin: 10px 0;
		padding: 5px 10px;
		border-radius: 2px;
		background: #383434;
		font-family: "Source Code Pro", monospace;
		overflow: auto;
		color: white;
		font-size: 13px;
		line-height: 1.6;
	}

	code {
		font-size: 90%;
		font-family: "Source Code Pro", monospace;
		padding: 2px 4px;
		border-radius: 2px;
		background: #383434;
		color: white;
		margin: 0 .2em;
	}
	
	i { font-variant: italic; }

	p.level-debug { color: #77aaea; }
	p.level-success { color: #66ab66; }
	p.level-info { color: white; }
	p.level-warn { color: #f9f67c; }
	p.level-error { color: #f14343; }
`
