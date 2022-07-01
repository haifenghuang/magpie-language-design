fn switchTest(name ) {
  switch name {
    case "welcome" {
       printf("Matched welcome: literal\n");
    }
    case /^Welcome$/ , /^WELCOME$/i {
       printf("Matched welcome: regular-expression\n");
    }
    case "Huang" + "HaiFeng" {
	printf("Matched HuangHaiFeng\n" );
    }
    case 3, 6, 9 {
        printf("Matched Number %d\n", name);
    }
    default {
	printf("Default case: %v\n", name );
    }
  }
}

switchTest( "welcome" );
switchTest( "WelCOME" );
switchTest( "HuangHaiFeng" );
switchTest( 3 );
switchTest( "Bob" );
switchTest( false );