package main

var defaultCSS = `<style>

table{
	border:solid;
}

`

var insertNewItemsCSS = defaultCSS + `

</style>`

var selectTodaysMenuCSS = defaultCSS + `
</style>`

var statsCSS = defaultCSS + `
.container{
	margin-bottom:25px
}
</style>`

var reopenSessionCSS = defaultCSS + `

#table{
	padding-top:10px
}

.btn{
	margin:7px
}



</style>`

var adminPageCSS = defaultCSS + `

button{
	margin:15px;
}
</style>`

var openSessionCSS = defaultCSS + `

#openSession{
	width: 23%;
	height: 200px;
	display: table;
	border-top: ridge;
	width:400px;
	margin-top:10px
}

p{
	font-size:20px
}

#data{
	font-size:20px;
}

b{
	margin-right:30px;
} 	

th{
	font-size:20px;
	border-bottom:solid;
	border-width:1px;
	padding:5px
}

td{
	font-size:17px;
	text-align:center
}

#bracelet{
	border-left:dotted;
	border-right:dotted;
	border-width:1px
}

#time{
	padding-right:10px
}

table{
	margin-bottom:25px
}

</style>`

var newSessionCSS = defaultCSS + `

#payment{
	padding-top: 5px
}

#last_session{
	width:200px;
	padding-bottom:15px
}

#button{
	padding-top: 20px;
}

p{
	font-size:20px;
	font-weight:bold
}

#checkboxes{
	font-size:20px
}


</style>`

var menuCSS = defaultCSS + `

#active{
	width:12px;
	height:12px;
	border:solid;
	border-width:.5px;
	display:inline-block;
	position:relative;
	top:5px;
	margin:5px
}

.items{
	text-decoration:none;
	color:#000;
	cursor:default;
}

html{
	height:100%
}

body{
	height:100%
}

th{
	padding:5px;
	text-align: left;
	border-bottom: dotted;
	border-right: dotted;
	border-width: 0.5px
}
  
tr{
	padding-right: 5px;
	border-bottom: dotted;
	border-right: dotted;
	border-width: 0.5px;
	text-align: right
}

td{
	padding: 5px;
	border-bottom: dotted;
	border-right: dotted;
	border-width: 0.5px;
	text-align: center
}

#lastOrder{
	width:200px;
	height:150px;
	float:left
}

#food{
	width:300px;
	height:150px;
	float:left;
}

#drink{
	width:200px;
	float:left;
	margin-left:-100px;
	margin-bottom:50px;
	height:150px
}

#misc{

	width:200px;
	height:150px;
	float:left;
}

div input h4{
	position:absolute;
	right:50%
}

#button{
	position:relative;
	top:-70px;
}

</style>`

var finalBillCSS = defaultCSS + `


#brclt{
	font-weight:bold;
	font-size:20px
}

button{
	margin-top:20px;
}

</style>`
