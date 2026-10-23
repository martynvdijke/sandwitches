// ignore_for_file: use_build_context_synchronously
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';

import '../../core/providers.dart';
import '../../core/roles.dart';

class RecipeCreateScreen extends ConsumerStatefulWidget {
  const RecipeCreateScreen({super.key});

  @override
  ConsumerState<RecipeCreateScreen> createState() => _RecipeCreateScreenState();
}

class _RecipeCreateScreenState extends ConsumerState<RecipeCreateScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleCtrl = TextEditingController();
  final _descriptionCtrl = TextEditingController();
  final _ingredientsCtrl = TextEditingController();
  final _instructionsCtrl = TextEditingController();
  final _servingsCtrl = TextEditingController(text: '1');
  final _priceCtrl = TextEditingController();
  final _tagsCtrl = TextEditingController();
  final ImagePicker _picker = ImagePicker();
  final ScrollController _scrollCtrl = ScrollController();

  String? _imagePath;
  String? _error;
  bool _saving = false;

  @override
  void dispose() {
    _titleCtrl.dispose();
    _descriptionCtrl.dispose();
    _ingredientsCtrl.dispose();
    _instructionsCtrl.dispose();
    _servingsCtrl.dispose();
    _priceCtrl.dispose();
    _tagsCtrl.dispose();
    _scrollCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickImage() async {
    final picked = await _picker.pickImage(
      source: ImageSource.gallery,
      imageQuality: 85,
      maxWidth: 1600,
    );
    if (picked == null || !mounted) return;
    setState(() => _imagePath = picked.path);
  }

  List<String> _parseTags() {
    return _tagsCtrl.text
        .split(',')
        .map((tag) => tag.trim())
        .where((tag) => tag.isNotEmpty)
        .toSet()
        .toList();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _saving = true;
      _error = null;
    });

    final priceText = _priceCtrl.text.trim();
    try {
      final recipe = await ref
          .read(repositoryProvider)
          .createRecipe(
            title: _titleCtrl.text.trim(),
            description: _descriptionCtrl.text.trim(),
            ingredients: _ingredientsCtrl.text.trim(),
            instructions: _instructionsCtrl.text.trim(),
            servings: int.parse(_servingsCtrl.text.trim()),
            price: priceText.isEmpty ? null : double.parse(priceText),
            tags: _parseTags(),
            imagePath: _imagePath,
          );

      ref.invalidate(recipesProvider);
      ref.invalidate(recipeOfDayProvider);
      ref.invalidate(tagsProvider);

      if (mounted) context.pushReplacement('/recipe/${recipe.id}');
    } catch (e) {
      if (mounted) setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final meAsync = ref.watch(meProvider);

    return meAsync.when(
      loading: () =>
          const Scaffold(body: Center(child: CircularProgressIndicator())),
      error: (e, _) => Scaffold(
        appBar: AppBar(title: const Text('New recipe')),
        body: Center(child: Text('Could not verify admin access: $e')),
      ),
      data: (me) {
        if (!isAdmin(me)) {
          return Scaffold(
            appBar: AppBar(title: const Text('New recipe')),
            body: const Center(
              child: Text(
                'Admins only — you do not have permission to add recipes.',
              ),
            ),
          );
        }
        return Scaffold(
          appBar: AppBar(title: const Text('New recipe')),
          body: Form(
            key: _formKey,
            child: SingleChildScrollView(
              key: const Key('recipe-form-scroll'),
              controller: _scrollCtrl,
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  if (_error != null) ...[
                    Card(
                      color: Theme.of(context).colorScheme.errorContainer,
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: Text(
                          _error!,
                          style: TextStyle(
                            color: Theme.of(context)
                                .colorScheme
                                .onErrorContainer,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                  ],
                  TextFormField(
                    key: const Key('recipe-title'),
                    controller: _titleCtrl,
                    textInputAction: TextInputAction.next,
                    decoration: const InputDecoration(
                      labelText: 'Title',
                      border: OutlineInputBorder(),
                    ),
                    validator: (v) => (v == null || v.trim().isEmpty)
                        ? 'Title is required'
                        : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _descriptionCtrl,
                    maxLines: 3,
                    decoration: const InputDecoration(
                      labelText: 'Description',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _ingredientsCtrl,
                    maxLines: 5,
                    decoration: const InputDecoration(
                      labelText: 'Ingredients',
                      helperText: 'One ingredient per line',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _instructionsCtrl,
                    maxLines: 5,
                    decoration: const InputDecoration(
                      labelText: 'Instructions',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: TextFormField(
                          key: const Key('recipe-servings'),
                          controller: _servingsCtrl,
                          keyboardType: TextInputType.number,
                          decoration: const InputDecoration(
                            labelText: 'Servings',
                            border: OutlineInputBorder(),
                          ),
                          validator: (v) {
                            final servings = int.tryParse(v?.trim() ?? '');
                            if (servings == null || servings < 1) {
                              return 'Must be 1 or more';
                            }
                            return null;
                          },
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: TextFormField(
                          key: const Key('recipe-price'),
                          controller: _priceCtrl,
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          decoration: const InputDecoration(
                            labelText: 'Price',
                            prefixText: '\$ ',
                            border: OutlineInputBorder(),
                          ),
                          validator: (v) {
                            final text = v?.trim() ?? '';
                            if (text.isEmpty) return null;
                            final price = double.tryParse(text);
                            if (price == null || price < 0) {
                              return 'Invalid price';
                            }
                            return null;
                          },
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    key: const Key('recipe-tags'),
                    controller: _tagsCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Tags',
                      helperText: 'Separate tags with commas',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Recipe photo',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  if (_imagePath != null) ...[
                    ClipRRect(
                      borderRadius: BorderRadius.circular(12),
                      child: Image.file(
                        File(_imagePath!),
                        height: 180,
                        width: double.infinity,
                        fit: BoxFit.cover,
                      ),
                    ),
                    const SizedBox(height: 8),
                  ],
                  Row(
                    children: [
                      OutlinedButton.icon(
                        onPressed: _saving ? null : _pickImage,
                        icon: const Icon(Icons.photo),
                        label: Text(
                          _imagePath == null ? 'Choose photo' : 'Change photo',
                        ),
                      ),
                      if (_imagePath != null) ...[
                        const SizedBox(width: 8),
                        TextButton.icon(
                          onPressed: _saving
                              ? null
                              : () => setState(() => _imagePath = null),
                          icon: const Icon(Icons.clear),
                          label: const Text('Remove'),
                        ),
                      ],
                    ],
                  ),
                  const SizedBox(height: 24),
                  FilledButton.icon(
                    onPressed: _saving ? null : _submit,
                    icon: _saving
                        ? const SizedBox(
                            height: 18,
                            width: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.add),
                    label: Text(_saving ? 'Creating…' : 'Create recipe'),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}
