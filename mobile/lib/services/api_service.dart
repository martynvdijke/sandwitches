import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/recipe.dart';

class ApiService {
  final String baseUrl;

  ApiService({required this.baseUrl});

  Future<List<Recipe>> fetchRecipes() async {
    final response = await http.get(Uri.parse('${baseUrl}v1/recipes'));

    if (response.statusCode == 200) {
      List jsonResponse = json.decode(response.body);
      return jsonResponse.map((data) => Recipe.fromJson(data)).toList();
    } else {
      throw Exception('Failed to load recipes');
    }
  }

  Future<Recipe> fetchRecipeDetail(int id) async {
    final response = await http.get(Uri.parse('${baseUrl}v1/recipes/$id'));

    if (response.statusCode == 200) {
      return Recipe.fromJson(json.decode(response.body));
    } else {
      throw Exception('Failed to load recipe detail');
    }
  }

  Future<Recipe?> fetchRecipeOfTheDay() async {
    final response =
        await http.get(Uri.parse('${baseUrl}v1/recipe-of-the-day'));

    if (response.statusCode == 200) {
      final body = response.body;
      if (body == 'null' || body.isEmpty) return null;
      return Recipe.fromJson(json.decode(body));
    } else {
      throw Exception('Failed to load recipe of the day');
    }
  }
}
