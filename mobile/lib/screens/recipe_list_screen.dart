import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/recipe_provider.dart';
import '../widgets/recipe_card.dart';

class RecipeListScreen extends StatefulWidget {
  const RecipeListScreen({super.key});

  @override
  State<RecipeListScreen> createState() => _RecipeListScreenState();
}

class _RecipeListScreenState extends State<RecipeListScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<RecipeProvider>().loadRecipes();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Sandwitches Recipes'),
        elevation: 0,
      ),
      body: Consumer<RecipeProvider>(
        builder: (context, provider, child) {
          if (provider.isLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (provider.errorMessage != null) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text('Error: ${provider.errorMessage}'),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => provider.loadRecipes(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (provider.recipes.isEmpty) {
            return const Center(child: Text('No recipes found.'));
          }

          return RefreshIndicator(
            onPressed: () => provider.loadRecipes(),
            child: ListView.builder(
              itemCount: provider.recipes.length,
              itemBuilder: (context, index) {
                return RecipeCard(recipe: provider.recipes[index]);
              },
            ),
          );
        },
      ),
    );
  }
}
