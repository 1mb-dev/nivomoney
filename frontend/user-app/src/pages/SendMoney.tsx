import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';

export function SendMoney() {
  const navigate = useNavigate();
  const { wallets, fetchWallets } = useWalletStore();
  const [sourceWalletId, setSourceWalletId] = useState('');
  const [destinationWalletId, setDestinationWalletId] = useState('');
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (wallets.length === 0) {
      fetchWallets().catch(err => console.error('Failed to fetch wallets:', err));
    }
  }, [wallets.length, fetchWallets]);

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!sourceWalletId) {
      newErrors.sourceWalletId = 'Please select a source wallet';
    }

    if (!destinationWalletId) {
      newErrors.destinationWalletId = 'Please enter destination wallet ID';
    } else if (destinationWalletId === sourceWalletId) {
      newErrors.destinationWalletId = 'Cannot send money to the same wallet';
    }

    if (!amount) {
      newErrors.amount = 'Please enter an amount';
    } else {
      const amountNum = parseFloat(amount);
      if (isNaN(amountNum) || amountNum <= 0) {
        newErrors.amount = 'Amount must be greater than 0';
      } else {
        const sourceWallet = wallets.find(w => w.id === sourceWalletId);
        if (sourceWallet && amountNum * 100 > sourceWallet.available_balance) {
          newErrors.amount = `Insufficient balance. Available: ${formatCurrency(sourceWallet.available_balance)}`;
        }
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(false);

    if (!validate()) return;

    setIsLoading(true);

    try {
      await api.createTransfer({
        source_wallet_id: sourceWalletId,
        destination_wallet_id: destinationWalletId,
        amount_paise: toPaise(parseFloat(amount)),
        currency: 'INR',
        description: description || 'Money transfer',
      });

      setSuccess(true);
      setSourceWalletId('');
      setDestinationWalletId('');
      setAmount('');
      setDescription('');

      // Refetch wallets to update balance
      await fetchWallets();

      // Navigate back to dashboard after 2 seconds
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send money');
      setIsLoading(false);
    }
  };

  const selectedWallet = wallets.find(w => w.id === sourceWalletId);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <h1 className="text-xl font-bold text-primary-600">Nivo</h1>
            <button onClick={() => navigate('/dashboard')} className="btn-secondary">
              Back to Dashboard
            </button>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="card">
          <h2 className="text-2xl font-bold mb-6">Send Money</h2>

          {error && (
            <div className="mb-4 p-4 bg-red-100 text-red-800 rounded-lg">
              {error}
            </div>
          )}

          {success && (
            <div className="mb-4 p-4 bg-green-100 text-green-800 rounded-lg">
              Transfer initiated successfully! Redirecting to dashboard...
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Source Wallet */}
            <div>
              <label htmlFor="sourceWallet" className="block text-sm font-medium text-gray-700 mb-2">
                From Wallet
              </label>
              <select
                id="sourceWallet"
                value={sourceWalletId}
                onChange={e => setSourceWalletId(e.target.value)}
                className="input-field"
                disabled={isLoading}
              >
                <option value="">Select a wallet</option>
                {wallets
                  .filter(w => w.status === 'active')
                  .map(wallet => (
                    <option key={wallet.id} value={wallet.id}>
                      {wallet.type.toUpperCase()} - {formatCurrency(wallet.available_balance)}
                    </option>
                  ))}
              </select>
              {errors.sourceWalletId && (
                <p className="text-sm text-red-600 mt-1">{errors.sourceWalletId}</p>
              )}
              {selectedWallet && (
                <p className="text-sm text-gray-600 mt-1">
                  Available balance: {formatCurrency(selectedWallet.available_balance)}
                </p>
              )}
            </div>

            {/* Destination Wallet */}
            <div>
              <label htmlFor="destinationWallet" className="block text-sm font-medium text-gray-700 mb-2">
                To Wallet ID
              </label>
              <input
                type="text"
                id="destinationWallet"
                value={destinationWalletId}
                onChange={e => setDestinationWalletId(e.target.value)}
                className="input-field"
                placeholder="Enter recipient's wallet ID"
                disabled={isLoading}
              />
              {errors.destinationWalletId && (
                <p className="text-sm text-red-600 mt-1">{errors.destinationWalletId}</p>
              )}
            </div>

            {/* Amount */}
            <div>
              <label htmlFor="amount" className="block text-sm font-medium text-gray-700 mb-2">
                Amount (₹)
              </label>
              <input
                type="number"
                id="amount"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                className="input-field"
                placeholder="0.00"
                step="0.01"
                min="0"
                disabled={isLoading}
              />
              {errors.amount && (
                <p className="text-sm text-red-600 mt-1">{errors.amount}</p>
              )}
            </div>

            {/* Description */}
            <div>
              <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
                Description (Optional)
              </label>
              <input
                type="text"
                id="description"
                value={description}
                onChange={e => setDescription(e.target.value)}
                className="input-field"
                placeholder="What's this for?"
                disabled={isLoading}
              />
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              className="btn-primary w-full"
              disabled={isLoading}
            >
              {isLoading ? 'Processing...' : 'Send Money'}
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}
