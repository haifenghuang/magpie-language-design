name = "Huang HaiFeng"
if name =~ /huang/i {
    println("Hello Huang");
}

if ( name !~ /xxx/ ) {
    println( "Hello xxx" );
}

if name =~ /Huang/ {
    println("Hello Huang");
}

regex = /\d+\t/.match("abc 123	mnj")
if (regex) {
    println("regex matched")
}

let arr = / /.split("ba na za")
if (len(arr) > 0) {
	printf("/ /.split('ba na za')[1]=%s\n", arr[1])
}
else {
	println("Not splitted")
}