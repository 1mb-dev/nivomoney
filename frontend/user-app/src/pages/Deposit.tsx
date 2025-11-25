import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../stores/walletStore';
import { api } from '../lib/api';
import { formatCurrency, toPaise } from '../lib/utils';

export function Deposit() {
  const navigate = useNavigate();
  const { wallets, fetchWallets } = useWalletStore();
  const [walletId, setWalletId] = useState('');
  const [amount, setAmount] = useState('');
  const [paymentMethod, setPaymentMethod] = useState('');
  const [reference, setReference] = useState('');
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

    if (!walletId) {
      newErrors.walletId = 'Please select a wallet';
    }

    if (!amount) {
      newErrors.amount = 'Please enter an amount';
    } else {
      const amountNum = parseFloat(amount);
      if (isNaN(amountNum) || amountNum <= 0) {
        newErrors.amount = 'Amount must be greater than 0';
      }
    }

    if (!paymentMethod) {
      newErrors.paymentMethod = 'Please select a payment method';
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
      await api.createDeposit({
        wallet_id: walletId,
        amount_paise: toPaise(parseFloat(amount)),
        description: `Deposit via ${paymentMethod}`,
        reference: reference || undefined,
      });

      setSuccess(true);
      setWalletId('');
      setAmount('');
      setPaymentMethod('');
      setReference('');

      // Refetch wallets to update balance
      await fetchWallets();

      // Navigate back to dashboard after 2 seconds
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to process deposit');
      setIsLoading(false);
    }
  };

  const selectedWallet = wallets.find(w => w.id === walletId);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <h1 className="text-xl font-bold text-primary-600">Nivo Money</h1>
            <button onClick={() => navigate('/dashboard')} className="btn-secondary">
              Back to Dashboard
            </button>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="card">
          <h2 className="text-2xl font-bold mb-6">Deposit Money</h2>

          {error && (
            <div className="mb-4 p-4 bg-red-100 text-red-800 rounded-lg">
              {error}
            </div>
          )}

          {success && (
            <div className="mb-4 p-4 bg-green-100 text-green-800 rounded-lg">
              Deposit initiated successfully! Redirecting to dashboard...
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Wallet Selection */}
            <div>
              <label htmlFor="wallet" className="block text-sm font-medium text-gray-700 mb-2">
                To Wallet
              </label>
              <select
                id="wallet"
                value={walletId}
                onChange={e => setWalletId(e.target.value)}
                className="input-field"
                disabled={isLoading}
              >
                <option value="">Select a wallet</option>
                {wallets
                  .filter(w => w.status === 'active')
                  .map(wallet => (
                    <option key={wallet.id} value={wallet.id}>
                      {wallet.type.toUpperCase()} - {formatCurrency(wallet.balance)}
                    </option>
                  ))}
              </select>
              {errors.walletId && (
                <p className="text-sm text-red-600 mt-1">{errors.walletId}</p>
              )}
              {selectedWallet && (
                <p className="text-sm text-gray-600 mt-1">
                  Current balance: {formatCurrency(selectedWallet.balance)}
                </p>
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

            {/* Payment Method */}
            <div>
              <label htmlFor="paymentMethod" className="block text-sm font-medium text-gray-700 mb-2">
                Payment Method
              </label>
              <select
                id="paymentMethod"
                value={paymentMethod}
                onChange={e => setPaymentMethod(e.target.value)}
                className="input-field"
                disabled={isLoading}
              >
                <option value="">Select payment method</option>
                <option value="bank_transfer">Bank Transfer</option>
                <option value="upi">UPI</option>
                <option value="card">Card</option>
                <option value="net_banking">Net Banking</option>
              </select>
              {errors.paymentMethod && (
                <p className="text-sm text-red-600 mt-1">{errors.paymentMethod}</p>
              )}
            </div>

            {/* Reference (Optional) */}
            <div>
              <label htmlFor="reference" className="block text-sm font-medium text-gray-700 mb-2">
                Reference Number (Optional)
              </label>
              <input
                type="text"
                id="reference"
                value={reference}
                onChange={e => setReference(e.target.value)}
                className="input-field"
                placeholder="Transaction reference"
                disabled={isLoading}
              />
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              className="btn-primary w-full"
              disabled={isLoading}
            >
              {isLoading ? 'Processing...' : 'Deposit Money'}
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}
