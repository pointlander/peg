import java.io.*;

public class TryWithResourcesDemo {
    public static void main(String[] args) {
        try (BufferedReader br = new BufferedReader(new FileReader("test.txt"))) {
            System.out.println(br.readLine());
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}

public class MultiCatchExample {
    public static void main(String[] args) {
        try {
            int a = 10 / 0;
            String str = null;
            str.length();
        } catch (ArithmeticException | NullPointerException e) {
            System.out.println("Exception caught: " + e.getMessage());
        }
    }
}

public class StringSwitchDemo {
    public static void main(String[] args) {
        String day = "MONDAY";
        switch (day) {
            case "MONDAY":
                System.out.println("Start of the workweek!");
                break;
            case "FRIDAY":
                System.out.println("Almost weekend!");
                break;
            default:
                System.out.println("Regular day.");
        }
    }
}

