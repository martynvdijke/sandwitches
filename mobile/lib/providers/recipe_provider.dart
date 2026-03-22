import 'package:flutter/material.dart';
import '../models/recipe.dart';
import '../services/api_service.dart';

class RecipeProvider with ChangeNotifier {
  final ApiService _apiService;
  List<Recipe> _recipes = [];
  Recipe? _featuredRecipe;
  bool _isLoading = false;
  String? _errorMessage;

  RecipeProvider({required ApiService apiService}) : _apiService = apiService;

  ApiService get apiService => _apiService;

  List<Recipe> get recipes => _recipes;
  Recipe? get featuredRecipe => _featuredRecipe;
  bool get isLoading => _isLoading;
  String? get errorMessage => _errorMessage;

  Future<void> loadRecipes() async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      _recipes = await _apiService.fetchRecipes();
    } catch (e) {
      _errorMessage = e.toString();
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> loadFeaturedRecipe() async {
    try {
      _featuredRecipe = await _apiService.fetchRecipeOfTheDay();
    } catch (e) {
      debugPrint("Error loading featured recipe: $e");
    } finally {
      notifyListeners();
    }
  }
}
