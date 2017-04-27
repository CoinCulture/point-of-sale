package main

import (
	"bytes"
	"html/template"
)

func renderTemplate(uselessString, theTemplate string, theStruct interface{}) ([]byte, error) {
	var err error
	t := template.New(uselessString)
	t, err = t.Parse(theTemplate)
	if err != nil {
		return []byte{}, err
	}
	var doc bytes.Buffer
	if err = t.Execute(&doc, theStruct); err != nil {
		return []byte{}, err
	}
	return doc.Bytes(), nil
}

var defaultHTML = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css">
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js"></script>
	<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js"></script>	

	<title></title>
</head>
<body>
	<nav class="navbar navbar-default">
		<div class="container-fluid">
			<div class="navbar-header">
				<a class="navbar-brand" href="/">MyBusiness</a>
			</div>
			<ul class="nav navbar-nav">
				<li active="active"><a href="/">Home</a></li>
				<li><a href="/newSession">New Session</a></li>
				<li><a href="/closeSession">Close Session</a></li>
				<li><a href="/addItems">Add Items</a></li>
				<li><a href="/adminPage">Admin</a></li>
			</ul>
		</div>
	</nav>
`

var openSessionsHTML = openSessionCSS + `
	
<h2>Open Sessions</h2>
<table>
	
	<tr>
		<th>Entry Time</th>
		<th>Locker #</th>
		<th>Running Total</th>
	</tr>
	
	{{range $index, $element := .}}
	
	<tr>
		<td id="time">{{ .EntryTime}}</td>
		<td id="bracelet">{{ .BraceletID}}</td>
		<td id="total">{{ .FinalBill.Total}}<br></td>
		{{end}}
	</tr>
	
</table>
`

var miscTemplate = `
<div id="misc">
	{{range .Miscs}} {{ .Name}} ${{ .Price}}<br>
		<input type="number" name="{{ .Name}}"><br>
	{{end}}
</div>
`

var lastSessionTemplate = `
<div id="last_session">
	<h3>Last Opened Session</h3>
	<table>
		<th>Bracelet #</th>
		<th>Owed</th>
		{{range $index, $element := .}}
		<tr>
			<td>{{ .BraceletID}}</td>
			<td>{{ .Total}}</td>
			{{end}}
		</tr>
	</table>
</div>`

func newSessionTemplate(lastOpenedSession, miscItemsForSale string) string {
	return defaultHTML + newSessionCSS + `
<div class="container">
<div class="row">
	<div class="col-md-6">
		<h1>New Session</h1>
		<p>Enter a locker number:</p>
		<form method="POST" action="/initializeSession">
			<input type="number" name="bracelet_id" autofocus><br>
			<p id="payment">Payment Method:</p>
			<div id="checkboxes">
			        <input type="radio" name="payment_method" value="general"> $5<br>
			       	<input type="radio" name="payment_method" value="punch_a_pass"> Punch<br>
			        <input type="radio" name="payment_method" value="pay_at_end"> Pay at end<br>
			        <input type="radio" name="payment_method" value="items_only"> Items Only (use locker # 0 only)<br>
			</div>
			<div id="button"><input type="submit" value="Confirm Entry"></div><br>	
				` + lastOpenedSession + openSessionsHTML + `

			</div>
			<div class="col-md-6">
				` + miscItemsForSale + `
			</div>
		</form>
	</div> 
</div>
</div>

</body>
</html>`
}

var menuTemplate = defaultHTML + menuCSS + `
<div class="container">
<h1>Select Items to Add</h1>

	<form  method="POST"action="/addItemsToASession">

		Locker Number:<br>
		<input id='input' type="number" name="bracelet_id" autofocus><div id="active"></div><br>
	
		<div id ="food">
			<h3>Food:</h3>
			{{range .Foods}}
			<a class="items" href="">{{ .Name}}</a>
			 ${{ .Price}}<br>
			<input type="number" name="{{ .Name}}"><br>
			{{end}}
		</div>

		<div id="drink">
			<h3>Drinks:</h3>
			{{range .Drinks}}
			<a class="items" href="">{{ .Name}}</a> 
			${{ .Price}}<br>
			<input type="number" name="{{ .Name}}"><br>
			{{end}}
		</div>

		<div id="misc">
			<h3>Miscellaneous:</h3>
			{{range .Miscs}}
			<a class="items" href="">{{ .Name}}</a>
			 ${{ .Price}}<br>
			<input type="number" name="{{ .Name}}"><br>
			{{end}}
		</div>

		<input id="button" type="submit" value="Add Items">

	</form>

	<div id="lastOrder">
		<h3>Last Order:</h3>

		<table>
			<tr>
				<th>Bracelet#</th>
				<th>Name</th>
				<th>Amount</th>
			</tr>
		{{range .LatestTransactions}}
			<tr>
				<td>{{ .BraceletID}}</td>
				<td>{{ .Name}}</td>
				<td>{{ .Amount}}</td>
			</tr>
		{{end}}
		</table>
	</div>
</div>

	<script>
	window.onload = function(){

	document.getElementById('input').addEventListener("keyup", getURL);

	function getURL(){
		var x = document.getElementById('input').value;
		var partialURL = "/isLockerActive?";
		var readyURL = partialURL.concat(x);
	
		function makeRequest(url){
			httpRequest = new XMLHttpRequest();
			if (!httpRequest){
			alert("could not create instance");
			return false
		}

		httpRequest.onreadystatechange = handleContents;				
		httpRequest.open('GET', readyURL);
		httpRequest.send()
	}
	
	function handleContents(){
		if (httpRequest.readyState === XMLHttpRequest.DONE){
			if (httpRequest.status === 200){
				if(httpRequest.responseText === "true"){
					document.getElementById("active").style.background = "#33cc33"
				}
				else{
					document.getElementById("active").style.background = "#FF0000"
				}
			}
		}
	}
	makeRequest("/isLockerActive")
	}

}
	</script>
	
</body>
</html>`

var displayInvoiceTemplate = defaultHTML + `
	<div class="container">
	<div class="row">
	<div class="col-md-12">

	<h1>Display Bill</h1>
	<form method="POST" action="/displayBill">
	<input type="number" name="bracelet_id" autofocus>
	<input type="submit" value="See Bill To Close">
	</form><br>
	<br>
	` + openSessionsHTML + `
	</div>
	</div>
	</div>
</body>
</html>`

var adminPageTemplate = defaultHTML + adminPageCSS + `
	<div class="container">
	<form method="POST" action="/selectTodaysMenuPage">
	<button type="submit" class="btn btn-default btn-block btn-lg">Select Today's Menu</button>
	</form>

	<form method="POST" action="/reopenOrDeleteSessionPage">
	<button type="submit" class="btn btn-default btn-block btn-lg">Re-open or Delete a Session</button>
	</form>

	<form method="POST" action="/endOfDayStatistics">
	<button type="submit" class="btn btn-default btn-block btn-lg">Close Day</button>
	</form>	

	<form method="POST" action="/insertNewItemsPage">
	<button type="submit" class="btn btn-default btn-block btn-lg">Manage Items</button>
	</form>

	<form method="POST" action="/statisticsOverview">
	<button type="submit" class="btn btn-default btn-block btn-lg">Statistics Overview</button>
	</form>
</div>	
</body>
</html>`

var endOfDayStatisticsTemplate = defaultHTML + statsCSS + `
	
<div class="container">
	<h2>End Of Day Overview</h2>
	<ul class="list-group">	
		<li class="list-group-item"><h3>Food Total: $ {{ .TotalFood}}</h3></li>
		<li class="list-group-item"><h3>Drinks Total: $ {{ .TotalDrink}}</h3></li>
		<li class="list-group-item"><h3>Miscellaneous Total: $ {{ .TotalMisc}}</h3></li>
		<li class="list-group-item"><h3>Grand Total: $ {{ .GrandTotal}}</h3></li>
		<li class="list-group-item"><h3>Number of Visits: {{ .NumberOfVisits}}</h2></li>
	</ul>

	<h4>Clicking this button will backup data & prepare the database for tomorrow</h4>

	<form method="POST" action="/closeDay">
		<input type="submit" value="Close Day" id="closeDay" />
	</form><br>

	<h2>Detailed Stats:</h2>
	<div class="row">
		<div class="col-md-6">
			<h3>Food</h3>	
			<table class="table table-bordered">
				<tr>
					<th>Item</th>
 					<th>Amount</th>
				</tr>
			{{range .Menu.Foods}}
				<tr>
					<td class="col-md-6">{{ .Name}}</td>
					<td class="col-md-6">{{ .Amount}}</td>
					{{end}}
				</tr>
			</table>
		</div>
	</div>

	<div class="row">
		<div class="col-md-6">	
			<h3>Drinks</h3>
			<table class="table table-bordered">
				<tr>
					<th>Item</th>
 					<th>Amount</th>
				</tr>
				{{range .Menu.Drinks}}
				<tr>
					<td class="col-md-6">{{ .Name}}</td>
					<td class="col-md-6">{{ .Amount}}</td>
					{{end}}
				</tr>
			</table>
		</div>
	</div>
	
	<div class="row">
		<div class="col-md-6">	
		<h3>Miscellaneous</h3>
		<table class="table table-bordered">
			<tr>
				<th>Item</th>
 				<th>Amount</th>
			</tr>
			{{range .Menu.Miscs}}
			<tr>
				<td class="col-md-6">{{ .Name}}</td>
				<td class="col-md-6">{{ .Amount}}</td>
				{{end}}
			</tr>
		</table>
	</div>
</div>

<script>

	document.getElementById("closeDay").addEventListener("click", createAlert);
	function createAlert(event){
		x = window.confirm("Are you sure?");
		if(x === false){
			event.preventDefault();
		}
	}

</script>

</body>

</html>`

var selectTodaysMenuPageTemplate = defaultHTML + defaultCSS + `
</style>
<div class="container">
	<h1>Set foods for the day</h1>

	<form method="POST" action="/selectTodaysMenu">
		{{range $index, $element := .}}
		<input type="checkbox" name="{{.Name}}" {{ .IsActive}}>
		{{ .Name}}<br>
		{{end}}
		<input type="submit" value="Update Foods">
	</form>
</div>
</body>
</html>`

var insertNewItemsTemplate = defaultHTML + defaultCSS + `
</style>
<div class="container">
	<h1>Add New Items to Menus</h1>
	<form method="POST" action="/insertNewItems">
		<div class="form-group">
			<label for="name">Name:</label>
			<input type="text" class="form-control" name="name">
			<p>If type is food, max length is 6 characters for printer formatting</p>
		</div>
	
		<div class="form-group">
			<label for="price">Price:</label>
			<input type="number" class="form-control" name="price">
		</div>

		<div class="form-group">
			<label for="type">Type:</label>
			<input type="text" class="form-control" name="type">
			<p>"food", "drink", "misc" ONLY</p>
		</div>
	
		<button type="submit" class="btn btn-default">Create New Item</button>

	</form>
</div>
</body>
</html>`

var statisticsOverviewTemplate = defaultHTML + defaultCSS + `
</style>
<div class="container">
	<h1>Get Statistics</h1>
	<form method="POST" action="/generateStatistics">
		<div class="form-group">
			<label for="beginning">Beginning Date:</label>
			<input type="text" class="form-control" name="beginning">
			<p>Use format 2017_06_23</p>
		</div>
	
		<div class="form-group">
			<label for="end">End Date:</label>
			<input type="text" class="form-control" name="end">
			<p>Use format 2017_07_23</p>
		</div>

		<button type="submit" class="btn btn-default">Get The Statistics</button>

	</form>
</div>
</body>
</html>`

var detailedStatisticsTemplate = defaultHTML + defaultCSS + `
</style>

<h1>Report for MyBusiness</h1>
<p>From: {{ .Beginning}}</p>
<p>To: {{ .End}}</p>

</div>
</body>
</html>`

var reopenOrDeleteSessionPageTemplate = defaultHTML + reopenSessionCSS + `

<div class="container">
<h2>Today's Closed Sessions</h2>

<div class="row" id="table">
	<div class="col-md-6">
		<table class="table table-condensed">
			<tr>
				<th>Entry Time</th>
				<th>Exit Time</th>
				<th>Locker #</th>
				<th>Total</th>
			</tr>
	
	{{range $index, $element := .}}
	
			<tr>
				<td>{{ .EntryTime}}</td>
				<td>todo</td>
				<td>{{ .BraceletID}}</td>
				<td>{{ .FinalBill.Total}}<br></td>
			{{end}}
			</tr>
		</table>
	</div>
</div>

<div class="row">
	<div class="col-md-12">
		<form method="POST" action="/reopenSession" class="form-inline">
  			<div class="form-group">	
				<label for="re-open">Enter locker # in closed sessions:</label>
			</div>
 			
			<div class="form-group">
				<input class="form-control" type="number" name="bracelet_id">
			</div>
			<button class="btn btn-default" type="submit">Re-open Session</button>
		</form>
	

		<form id="delete" method="POST" action="/deleteSession" class="form-inline">
  			<div class="form-group">
				<label for="delete">Enter locker # to delete a session:</label>
			</div>

			<div class="form-group">
				<input class="form-control"  type="number" name="bracelet_id">
			</div>
			<button class="btn btn-default" type="submit">Delete Session</button>
		</form>
	</div>

</div>

</div>

<script>

	var btn = document.getElementsByTagName("button");
	btn[0].addEventListener("click", reopenSess);
	btn[1].addEventListener("click", deleteSess);
	
	function reopenSess(event){
		x = window.confirm("Are you sure you want to RE-OPEN this session?");
		if(x === false){
			event.preventDefault()
		}
	}
	function deleteSess(event){
		x = window.confirm("Are you sure you want to DELETE this session?");
                if(x === false){
                        event.preventDefault()
                }
	}

</script>

</body>
</html>`

var finalBillTemplate = defaultHTML + finalBillCSS + `
<div class="container">
	<h1>Final Bill</h1>
	<div class="row">
		<div class="col-md-6">
			<ul class="list-group">
				<li class="list-group-item"><span id="brclt">Bracelet # {{ .BraceletID}}</span></li>
				<li class="list-group-item">Admission Type: {{ .AdmissionType}}</li>
				<li class="list-group-item">Invoice #: {{ .InvoiceID}}</li>
				<li class="list-group-item">Entry Time: {{ .EntryTime}}</li>
			</ul>
		</div>
	</div>

	<div class="row">
		<div class="col-md-6">
			<table class="table table-condensed">
				<h3>Food:</h3>
				<tr>
					<th>Name</th>
					<th>Amount</th>
					<th>Total</th>
				</tr>
		</div>
	</div>
				{{range .FinalBill.Foods}}	

				<tr>
					<td>{{ .Name}}</td>
					<td>{{ .Amount}}</td>
					<td>{{ .Total}}</td>
				</tr>

				{{end}}
	
			</table><br>

	<table class="table table-condensed">
		<h3>Drinks:</h3>
		<tr>
			<th>Name</th>
			<th>Amount</th>
			<th>Total</th>
		</tr>

		{{range .FinalBill.Drinks}}

		<tr>
			<td>{{ .Name}}</td>
			<td>{{ .Amount}}</td>
			<td>{{ .Total}}</td>
		</tr>

		{{end}}

	</table><br>
	
	<table class="table table-condensed">
		<h3>Miscellaneous:</h3>
		<tr>
			<th>Name</th>
			<th>Amount</th>
			<th>Total</th>
		</tr>
	
		{{range .FinalBill.Miscs}}

		<tr>
			<td>{{ .Name}}</td>
			<td>{{ .Amount}}</td>
			<td>{{ .Total}}</td>
		</tr>

		{{end}}

	</table><br>

	<table class="table table-condensed">
		<h3>Total:</h3>
		<th>
			{{ .FinalBill.Total}}
		</th>
	</table>

	<form method="POST" action="/closeBill?{{ .BraceletID}}&{{ .InvoiceID}}&{{ .FinalBill.Total}}">
		<button type="submit" class="btn btn-default">Close Bill</button>
	</form>
</div>
</body>
</html>`
