import 'package:flutter_test/flutter_test.dart';
import 'package:sandwitches/core/models.dart';

void main(){
  test('ingredient scaling mapping', (){
    final json = [
      {'original_line':'1 cup milk','scaled_line':'2 cups milk','quantity':2,'unit':'cups','name':'milk'},
      {'original_line':'3 eggs','scaled_line':'6 eggs','quantity':6,'unit':'','name':'eggs'},
    ];
    final list = json.map((e)=> ScaledIngredient.fromJson(e)).toList();
    expect(list.length, 2);
    expect(list[0].scaledLine, '2 cups milk');
    expect(list[1].quantity, 6);
  });
  test('thumb url edge cases', (){
    expect(buildThumbUrl('http://localhost:8000', '/media/recipes/a.jpg'), 'http://localhost:8000/thumb/recipes/a.jpg?w=800');
    expect(buildThumbUrl('http://localhost:8000/', 'media/recipes/a.jpg', w:200), 'http://localhost:8000/thumb/recipes/a.jpg?w=200');
    expect(buildThumbUrl('', '/media/recipes/a.jpg'), '/thumb/recipes/a.jpg?w=800');
  });
}
