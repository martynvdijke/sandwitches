import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:dio/dio.dart';
import 'package:sandwitches/core/providers.dart';
import 'package:sandwitches/core/models.dart';
import 'package:sandwitches/core/roles.dart';
import 'package:sandwitches/features/home/home_screen.dart';
import 'package:sandwitches/features/recipe/recipe_create_screen.dart';
import 'package:sandwitches/features/recipe/recipe_detail_screen.dart';

void main() {
  test('isAdmin recognizes the backend staff flags', () {
    expect(isAdmin(null), isFalse);
    expect(isAdmin(<String, dynamic>{}), isFalse);
    expect(
      isAdmin(<String, dynamic>{'is_staff': false, 'is_superuser': false}),
      isFalse,
    );
    expect(isAdmin(<String, dynamic>{'is_staff': true}), isTrue);
    expect(isAdmin(<String, dynamic>{'is_superuser': true}), isTrue);
  });

  testWidgets('home shows the add recipe action only for admins', (
    tester,
  ) async {
    final repo = FakeAdminRepo([
      Recipe(
        id: 1,
        title: 'Existing',
        slug: 'existing',
        description: '',
        ingredients: '',
        instructions: '',
        servings: 1,
        price: '1',
        image: '',
      ),
    ]);

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          repositoryProvider.overrideWithValue(repo),
          meProvider.overrideWith(
            (ref) async => <String, dynamic>{'is_staff': true},
          ),
        ],
        child: const MaterialApp(home: HomeScreen()),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.byType(FloatingActionButton), findsOneWidget);
    expect(find.text('Add recipe'), findsOneWidget);
  });

  testWidgets('home hides the add recipe action for non-admins', (
    tester,
  ) async {
    final repo = FakeAdminRepo([]);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          repositoryProvider.overrideWithValue(repo),
          meProvider.overrideWith(
            (ref) async => <String, dynamic>{'is_staff': false},
          ),
        ],
        child: const MaterialApp(home: HomeScreen()),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.byType(FloatingActionButton), findsNothing);
  });

  testWidgets('create form validates the title', (tester) async {
    final repo = FakeAdminRepo([]);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          repositoryProvider.overrideWithValue(repo),
          meProvider.overrideWith(
            (ref) async => <String, dynamic>{'is_staff': true},
          ),
        ],
        child: MaterialApp.router(routerConfig: _router()),
      ),
    );
    await tester.pumpAndSettle();

    final controller = _formScroll(tester);
    controller.jumpTo(controller.position.maxScrollExtent);
    await tester.pump();
    await tester.tap(find.text('Create recipe'));
    await tester.pump();
    _formScroll(tester).jumpTo(0);
    await tester.pumpAndSettle();
    expect(find.text('Title is required'), findsOneWidget);
    expect(repo.createdTitle, isNull);
  });

  testWidgets('create form submits and opens the new recipe', (tester) async {
    final repo = FakeAdminRepo([]);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          repositoryProvider.overrideWithValue(repo),
          meProvider.overrideWith(
            (ref) async => <String, dynamic>{'is_staff': true},
          ),
        ],
        child: MaterialApp.router(routerConfig: _router()),
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byKey(const Key('recipe-title')),
      'Admin Sandwich',
    );
    await tester.enterText(
      find.byKey(const Key('recipe-tags')),
      'quick, lunch, quick',
    );
    FocusManager.instance.primaryFocus?.unfocus();
    await tester.pumpAndSettle();

    final controller = _formScroll(tester);
    controller.jumpTo(controller.position.maxScrollExtent);
    await tester.pump();
    await tester.tap(find.text('Create recipe'));
    await tester.pumpAndSettle();

    expect(repo.createdTitle, 'Admin Sandwich');
    expect(repo.createdTags, ['quick', 'lunch']);
    expect(find.text('Admin Sandwich'), findsOneWidget);
  });

  testWidgets('create form shows API errors', (tester) async {
    final repo = FakeAdminRepo([])..throwError = true;
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          repositoryProvider.overrideWithValue(repo),
          meProvider.overrideWith(
            (ref) async => <String, dynamic>{'is_staff': true},
          ),
        ],
        child: MaterialApp.router(routerConfig: _router()),
      ),
    );
    await tester.pumpAndSettle();

    await tester.enterText(find.byKey(const Key('recipe-title')), 'Denied');
    FocusManager.instance.primaryFocus?.unfocus();
    await tester.pumpAndSettle();
    final controller = _formScroll(tester);
    controller.jumpTo(controller.position.maxScrollExtent);
    await tester.pump();
    await tester.tap(find.text('Create recipe'));
    await tester.pumpAndSettle();
    _formScroll(tester).jumpTo(0);
    await tester.pumpAndSettle();

    expect(find.text('Only admins can create recipes'), findsOneWidget);
  });

  testWidgets('deep link to the create form is blocked for non-admins', (
    tester,
  ) async {
    final repo = FakeAdminRepo([]);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          repositoryProvider.overrideWithValue(repo),
          meProvider.overrideWith(
            (ref) async => <String, dynamic>{'is_staff': false},
          ),
        ],
        child: MaterialApp.router(routerConfig: _router()),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.textContaining('Admins only'), findsOneWidget);
  });
}

ScrollController _formScroll(WidgetTester tester) {
  return tester
      .widget<SingleChildScrollView>(
        find.byKey(const Key('recipe-form-scroll')),
      )
      .controller!;
}

GoRouter _router() {
  return GoRouter(
    initialLocation: '/admin/recipes/new',
    routes: [
      GoRoute(
        path: '/admin/recipes/new',
        builder: (_, _) => const RecipeCreateScreen(),
      ),
      GoRoute(
        path: '/recipe/:id',
        builder: (_, state) => RecipeDetailScreen(
          recipeId: int.parse(state.pathParameters['id']!),
        ),
      ),
    ],
  );
}

class FakeAdminRepo extends ApiRepository {
  final List<Recipe> recipes;
  String? createdTitle;
  String? createdImagePath;
  List<String>? createdTags;
  bool throwError = false;

  FakeAdminRepo(this.recipes) : super(Dio(), 'https://example.com');

  @override
  Future<List<Recipe>> fetchRecipes({
    String? search,
    String? tag,
    String? isApproved,
    int? page,
    int? perPage,
  }) async {
    return recipes;
  }

  @override
  Future<Recipe?> fetchRecipeOfDay() async => null;

  @override
  Future<List<Tag>> fetchTags() async => [];

  @override
  Future<Recipe> fetchRecipe(int id) async {
    return recipes.firstWhere(
      (recipe) => recipe.id == id,
      orElse: () => recipes.first,
    );
  }

  @override
  Future<List<ScaledIngredient>> scaleIngredients(int id, int target) async =>
      [];

  @override
  Future<Map<String, dynamic>> fetchRatingSummary(int id) async => {
    'average': 0,
    'count': 0,
  };

  @override
  Future<List<Rating>> fetchRatings(int id) async => [];

  @override
  Future<Recipe> createRecipe({
    required String title,
    String description = '',
    String ingredients = '',
    String instructions = '',
    int servings = 1,
    double? price,
    List<String> tags = const [],
    String? imagePath,
  }) async {
    if (throwError) throw ApiException('Only admins can create recipes');
    createdTitle = title;
    createdImagePath = imagePath;
    createdTags = tags;
    final recipe = Recipe(
      id: 99,
      title: title,
      slug: 'admin-sandwich',
      description: description,
      ingredients: ingredients,
      instructions: instructions,
      servings: servings,
      price: price?.toString() ?? '0',
      image: imagePath ?? '',
      tags: [],
    );
    recipes.add(recipe);
    return recipe;
  }
}
