import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/config_provider.dart';

class SetupScreen extends StatefulWidget {
  const SetupScreen({super.key});

  @override
  State<SetupScreen> createState() => _SetupScreenState();
}

class _SetupScreenState extends State<SetupScreen> {
  final _urlController = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  @override
  void dispose() {
    _urlController.dispose();
    super.dispose();
  }

  void _submit() {
    if (_formKey.currentState!.validate()) {
      context.read<ConfigProvider>().saveApiUrl(_urlController.text);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(32.0),
          child: Form(
            key: _formKey,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Icon(
                  Icons.lunch_dining,
                  size: 100,
                  color: Colors.orange,
                ),
                const SizedBox(height: 32),
                Text(
                  'Welcome to Sandwitches',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: Colors.orange[800],
                      ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 16),
                const Text(
                  'Please enter the URL of your Sandwitches instance to continue.',
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 32),
                TextFormField(
                  controller: _urlController,
                  decoration: const InputDecoration(
                    labelText: 'Instance URL',
                    hintText: 'https://sandwitches.example.com',
                    border: OutlineInputBorder(),
                    prefixIcon: Icon(Icons.link),
                  ),
                  keyboardType: TextInputType.url,
                  validator: (value) {
                    if (value == null || value.trim().isEmpty) {
                      return 'Please enter a URL';
                    }
                    if (!value.startsWith('http://') &&
                        !value.startsWith('https://')) {
                      return 'URL must start with http:// or https://';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 24),
                ElevatedButton(
                  onPressed: _submit,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.orange[800],
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  child: const Text('Connect'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
